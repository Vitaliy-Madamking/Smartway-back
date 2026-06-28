server  - Точка входа

loadtest - Утилита тестирования

internal/
 domain - Сущности
 usecase/     Бизнес-логика
 delivery/    HTTP-обработчики
 pkg/         Алгоритмы
 repository/ Хранение
 data/           Тестовые данные



// запуск (Для Ромы)(качаешь файлы и устанавливаешь все в 1 папку под названием hotel-matcher)
go mod tidy
go run cmd/server/main.go
