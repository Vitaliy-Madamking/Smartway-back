package domain

// Hotel — бизнес-сущность отеля
type Hotel struct {
	ID        string
	Source    string
	Name      string
	Address   string
	City      string
	Country   string
	Latitude  float64
	Longitude float64
}

// Validate проверяет обязательные поля
func (h Hotel) Validate() error {
	if h.ID == "" {
		return ErrEmptyID
	}
	if h.Name == "" {
		return ErrEmptyName
	}
	return nil
}