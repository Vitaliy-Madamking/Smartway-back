package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hotel-matcher/internal/domain"
)

type GroupRepository struct {
	db *pgxpool.Pool
}

func NewGroupRepository(db *pgxpool.Pool) *GroupRepository {
	return &GroupRepository{db: db}
}

// GetAll возвращает все группы со списком отелей внутри каждой
func (r *GroupRepository) GetAll(ctx context.Context) ([]domain.Group, error) {
	const query = `
		SELECT
			g.id,
			g.primary_name,
			g.confidence,
			g.match_score,
			g.providers_count,
			g.hotels_count,
			g.match_reasons,
			g.created_at,
			COALESCE(
				json_agg(json_build_object(
					'id',               h.id,
					'supplier_name',    h.supplier_name,
					'supplier_hotel_id',h.supplier_hotel_id,
					'name',             h.name,
					'address',          h.address,
					'city',             h.city,
					'country',          h.country,
					'country_code',     h.country_code,
					'latitude',         h.latitude,
					'longitude',        h.longitude,
					'stars',            h.stars
				)) FILTER (WHERE h.id IS NOT NULL),
				'[]'
			) AS hotels
		FROM hotel_groups g
		LEFT JOIN hotel_group_members m ON m.group_id = g.id
		LEFT JOIN hotels h ON h.id = m.hotel_id
		GROUP BY g.id
		ORDER BY g.id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return groups, nil
}

// GetByID возвращает группу по id со списком отелей
func (r *GroupRepository) GetByID(ctx context.Context, id int64) (*domain.Group, error) {
	const query = `
		SELECT
			g.id,
			g.primary_name,
			g.confidence,
			g.match_score,
			g.providers_count,
			g.hotels_count,
			g.match_reasons,
			g.created_at,
			COALESCE(
				json_agg(json_build_object(
					'id',               h.id,
					'supplier_name',    h.supplier_name,
					'supplier_hotel_id',h.supplier_hotel_id,
					'name',             h.name,
					'address',          h.address,
					'city',             h.city,
					'country',          h.country,
					'country_code',     h.country_code,
					'latitude',         h.latitude,
					'longitude',        h.longitude,
					'stars',            h.stars
				)) FILTER (WHERE h.id IS NOT NULL),
				'[]'
			) AS hotels
		FROM hotel_groups g
		LEFT JOIN hotel_group_members m ON m.group_id = g.id
		LEFT JOIN hotels h ON h.id = m.hotel_id
		WHERE g.id = $1
		GROUP BY g.id
	`

	row := r.db.QueryRow(ctx, query, id)
	g, err := scanGroup(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, fmt.Errorf("get group %d: %w", id, err)
	}

	return &g, nil
}

// SaveResult сохраняет результат матчинга: группы + связи с отелями
func (r *GroupRepository) SaveResult(ctx context.Context, result *domain.Result) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	const insertGroup = `
		INSERT INTO hotel_groups
			(primary_name, confidence, match_score, providers_count, hotels_count, match_reasons)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	const insertMember = `
		INSERT INTO hotel_group_members (hotel_id, group_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	for _, group := range result.Groups {
		// Определяем primary_name — имя первого отеля в группе
		primaryName := ""
		if len(group.Hotels) > 0 {
			primaryName = group.Hotels[0].Name
		}

		// Собираем причины матчинга (уникальные supplier-имена)
		reasons := buildMatchReasons(group.Hotels)
		reasonsJSON, err := json.Marshal(reasons)
		if err != nil {
			return fmt.Errorf("marshal match_reasons: %w", err)
		}

		// Считаем уникальных провайдеров
		providers := countUniqueProviders(group.Hotels)

		var groupID int64
		err = tx.QueryRow(ctx, insertGroup,
			primaryName,
			group.ConfidenceScore,
			group.ConfidenceScore, // match_score = confidence пока одно и то же
			providers,
			len(group.Hotels),
			reasonsJSON,
		).Scan(&groupID)
		if err != nil {
			return fmt.Errorf("insert group: %w", err)
		}

		// Привязываем отели к группе
		for _, hotel := range group.Hotels {
			if _, err := tx.Exec(ctx, insertMember, hotel.ID, groupID); err != nil {
				return fmt.Errorf("insert member hotel_id=%s group_id=%d: %w", hotel.ID, groupID, err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// --- helpers ---

// scanRow — общий интерфейс для pgx.Row и pgx.Rows
type scanRow interface {
	Scan(dest ...any) error
}

func scanGroup(row scanRow) (domain.Group, error) {
	var g domain.Group
	var hotelsJSON []byte

	if err := row.Scan(
		&g.ID,
		&g.PrimaryName,
		&g.ConfidenceScore,
		&g.MatchScore,
		&g.ProvidersCount,
		&g.HotelsCount,
		&g.MatchReasons,
		&g.CreatedAt,
		&hotelsJSON,
	); err != nil {
		return g, fmt.Errorf("scan group: %w", err)
	}

	if err := json.Unmarshal(hotelsJSON, &g.Hotels); err != nil {
		return g, fmt.Errorf("unmarshal hotels: %w", err)
	}

	return g, nil
}

func buildMatchReasons(hotels []domain.Hotel) map[string]any {
	suppliers := make([]string, 0, len(hotels))
	seen := make(map[string]bool)
	for _, h := range hotels {
		if !seen[h.SupplierName] {
			seen[h.SupplierName] = true
			suppliers = append(suppliers, h.SupplierName)
		}
	}
	return map[string]any{
		"matched_suppliers": suppliers,
		"total":             len(hotels),
	}
}

func countUniqueProviders(hotels []domain.Hotel) int {
	seen := make(map[string]bool)
	for _, h := range hotels {
		seen[h.SupplierName] = true
	}
	return len(seen)
}
