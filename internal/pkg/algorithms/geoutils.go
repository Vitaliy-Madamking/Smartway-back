package algorithms

import "math"

// radiusEarth - радиус Земли в километрах (используется в формуле гаверсинусов)
const earthRadius = 6371.0

// Haversine вычисляет расстояние между двумя точками на сфере по их координатам.
// Возвращает расстояние в километрах.
func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Переводим градусы в радианы (математические функции Go работают с радианами)
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Разница координат
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	// Формула гаверсинусов:
	// a = sin²(Δlat/2) + cos(lat1) * cos(lat2) * sin²(Δlon/2)
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	// c = 2 * atan2(√a, √(1-a)) — центральный угол в радианах
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Расстояние = радиус * угол
	return earthRadius * c
}

// CompareCoordinates преобразует расстояние между отелями в оценку схожести (0..1).
// Чем ближе отели, тем выше оценка.
func CompareCoordinates(lat1, lon1, lat2, lon2 float64) float64 {
	// Если координаты не указаны (0,0) — возвращаем нейтральное значение 0.5
	// Это не штрафует отели без координат
	if lat1 == 0 || lon1 == 0 || lat2 == 0 || lon2 == 0 {
		return 0.5
	}

	// Получаем расстояние в километрах
	dist := Haversine(lat1, lon1, lat2, lon2)

	// Преобразуем расстояние в оценку:
	// < 0.5 км  → 1.0  (отлично, отели рядом)
	// < 2.0 км  → 0.8  (хорошо, один район)
	// < 5.0 км  → 0.5  (средне, один город)
	// < 10.0 км → 0.3  (слабо, разные концы города)
	// >= 10 км  → 0.0  (нет совпадения, разные города)
	switch {
	case dist < 0.5:
		return 1.0
	case dist < 2.0:
		return 0.8
	case dist < 5.0:
		return 0.5
	case dist < 10.0:
		return 0.3
	default:
		return 0.0
	}
}