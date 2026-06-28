package usecase

import (
	"context"
	"hotel-matcher/internal/domain"
)

type Matcher interface {
	Match(ctx context.Context, hotels []domain.Hotel, cfg domain.Config) (*domain.Result, error)
}

type HotelRepository interface {
	Save(ctx context.Context, hotel domain.Hotel) error
	GetByID(ctx context.Context, id string) (*domain.Hotel, error)
	GetAll(ctx context.Context) ([]domain.Hotel, error)
}