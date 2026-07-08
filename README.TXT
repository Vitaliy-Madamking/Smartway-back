server  - Точка входа

loadtest - Утилита тестирования

internal/
 domain - Сущности
 usecase/     Бизнес-логика
 delivery/    HTTP-обработчики
 pkg/         Алгоритмы
 repository/ Хранение
 data/           Тестовые данные



go mod tidy
go run cmd/server/main.go
