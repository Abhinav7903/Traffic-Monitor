package models

type GeoRecord struct {
	IP           string  `json:"ip"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	Region       string  `json:"region"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	PostalCode   string  `json:"postal_code"`
	ASN          uint    `json:"asn"`
	Organization string  `json:"organization"`
}
