package domain

import "errors"

var (
	ErrEmptyID       = errors.New("id отеля не может быть пустым")
	ErrEmptyName     = errors.New("Название отеля не может быть пустым")
	ErrNoHotels      = errors.New("Отели не указаны для сопоставления")
	ErrInvalidConfig = errors.New("Неверная конфигурация")
	ErrNotFound      = errors.New("Отель не найден")
	ErrGroupNotFound = errors.New("Группа не найдена")
)
