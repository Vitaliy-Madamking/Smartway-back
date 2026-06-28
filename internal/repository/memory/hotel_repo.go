package memory

import (
	"context"
	"hotel-matcher/internal/domain"
	"sync"
)

type HotelRepository struct {
	mu     sync.RWMutex
	hotels map[string]domain.Hotel
}

func NewHotelRepository() *HotelRepository {
	return &HotelRepository{hotels: make(map[string]domain.Hotel)}
}

func (r *HotelRepository) Save(ctx context.Context, hotel domain.Hotel) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hotels[hotel.ID] = hotel
	return nil
}

func (r *HotelRepository) GetByID(ctx context.Context, id string) (*domain.Hotel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.hotels[id]
	if !ok {
		return nil, nil
	}
	return &h, nil
}

func (r *HotelRepository) GetAll(ctx context.Context) ([]domain.Hotel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Hotel, 0, len(r.hotels))
	for _, h := range r.hotels {
		res = append(res, h)
	}
	return res, nil
}