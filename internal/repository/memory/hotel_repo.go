package memory

import (
	"context"
	"hotel-matcher/internal/domain"
	"sync"
)

// HotelRepository — in-memory реализация репозитория
// Хранит отели в map и защищает доступ мьютексом
type HotelRepository struct {
	mu     sync.RWMutex              // защита от параллельного доступа
	hotels map[string]domain.Hotel   // хранилище: ID отеля → данные отеля
}

// NewHotelRepository — конструктор, создаёт пустое хранилище
func NewHotelRepository() *HotelRepository {
	return &HotelRepository{hotels: make(map[string]domain.Hotel)}
}

// Save — сохраняет или обновляет отель
// Потокобезопасно: Lock() для записи
func (r *HotelRepository) Save(ctx context.Context, hotel domain.Hotel) error {
	r.mu.Lock()         // блокируем для записи (эксклюзивно)
	defer r.mu.Unlock() // разблокируем после завершения
	r.hotels[hotel.ID] = hotel
	return nil
}

// GetByID — получает отель по ID
// Потокобезопасно: RLock() для чтения (можно параллельно с другими чтениями)
func (r *HotelRepository) GetByID(ctx context.Context, id string) (*domain.Hotel, error) {
	r.mu.RLock()         // блокируем для чтения
	defer r.mu.RUnlock() // разблокируем после завершения
	h, ok := r.hotels[id]
	if !ok {
		return nil, nil // отель не найден
	}
	return &h, nil
}

// GetAll — возвращает все отели в виде слайса
// Потокобезопасно: RLock() для чтения
func (r *HotelRepository) GetAll(ctx context.Context) ([]domain.Hotel, error) {
	r.mu.RLock()         // блокируем для чтения
	defer r.mu.RUnlock() // разблокируем после завершения
	// Создаём слайс с ёмкостью = количеству отелей (оптимизация)
	res := make([]domain.Hotel, 0, len(r.hotels))
	for _, h := range r.hotels {
		res = append(res, h)
	}
	return res, nil
}