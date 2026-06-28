package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"hotel-matcher/internal/domain"
)

type HotelRepository struct {
	db *pgxpool.Pool
}

func NewHotelRepository(db *pgxpool.Pool) *HotelRepository {
	return &HotelRepository{db: db}
}

// GetAll возвращает все отели из таблицы hotels
func (r *HotelRepository) GetAll(ctx context.Context) ([]domain.Hotel, error) {
	const query = `
		SELECT
			id, supplier_name, supplier_hotel_id,
			name, address, city, country, country_code,
			latitude, longitude, stars
		FROM hotels
		ORDER BY id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query hotels: %w", err)
	}
	defer rows.Close()

	var hotels []domain.Hotel
	for rows.Next() {
		var h domain.Hotel
		if err := rows.Scan(
			&h.ID,
			&h.SupplierName,
			&h.SupplierHotelID,
			&h.Name,
			&h.Address,
			&h.City,
			&h.Country,
			&h.CountryCode,
			&h.Latitude,
			&h.Longitude,
			&h.Stars,
		); err != nil {
			return nil, fmt.Errorf("scan hotel: %w", err)
		}
		hotels = append(hotels, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return hotels, nil
}

// SaveBatch сохраняет отели пачкой, игнорируя дубли по (supplier_name, supplier_hotel_id)
// Возвращает сохранённые отели с присвоенными id
func (r *HotelRepository) SaveBatch(ctx context.Context, hotels []domain.Hotel) ([]domain.Hotel, error) {
	const query = `
		INSERT INTO hotels
			(supplier_name, supplier_hotel_id, name, address, city, country, country_code, latitude, longitude, stars)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (supplier_name, supplier_hotel_id) DO UPDATE
			SET name        = EXCLUDED.name,
			    address     = EXCLUDED.address,
			    city        = EXCLUDED.city,
			    country     = EXCLUDED.country,
			    country_code= EXCLUDED.country_code,
			    latitude    = EXCLUDED.latitude,
			    longitude   = EXCLUDED.longitude,
			    stars       = EXCLUDED.stars
		RETURNING id
	`

	saved := make([]domain.Hotel, 0, len(hotels))

	// pgx не поддерживает batch RETURNING напрямую через CopyFrom,
	// поэтому используем транзакцию + prepared statement
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, h := range hotels {
		var id string
		err := tx.QueryRow(ctx, query,
			h.SupplierName, h.SupplierHotelID,
			h.Name, h.Address, h.City, h.Country, h.CountryCode,
			h.Latitude, h.Longitude, h.Stars,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("insert hotel %q: %w", h.Name, err)
		}
		h.ID = id
		saved = append(saved, h)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return saved, nil
}
