package domain

// Hotel — бизнес-сущность отеля
type Hotel struct {
	ID              string  `json:"id"`
	SupplierName    string  `json:"supplier_name"`
	SupplierHotelID string  `json:"supplier_hotel_id"`
	Name            string  `json:"name"`
	Address         string  `json:"address"`
	City            string  `json:"city"`
	Country         string  `json:"country"`
	CountryCode     string  `json:"country_code"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Stars           int     `json:"stars"`
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
