package usecase

import (
	"context"

	"hotel-matcher/internal/domain"
)

// Matcher — основной интерфейс матчинга
type Matcher interface {
	Match(ctx context.Context, hotels []domain.Hotel, cfg domain.Config) (*domain.Result, error)
}

// HotelRepository — полный интерфейс (используется в usecase)
type HotelRepository interface {
	HotelReader
}

// HotelReader — интерфейс для чтения и записи отелей (инжектируется в handler)
type HotelReader interface {
	GetAll(ctx context.Context) ([]domain.Hotel, error)
	SaveBatch(ctx context.Context, hotels []domain.Hotel) ([]domain.Hotel, error)
}

// GroupReader — интерфейс для чтения и записи групп (инжектируется в handler)
type GroupReader interface {
	GetAll(ctx context.Context) ([]domain.Group, error)
	GetByID(ctx context.Context, id int64) (*domain.Group, error)
	SaveResult(ctx context.Context, result *domain.Result) error
}
