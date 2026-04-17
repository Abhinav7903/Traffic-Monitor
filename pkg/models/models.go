package models

type GeoRecord struct {
	IP           string  `json:"ip"`
	Country      string  `json:"country"`
	CountryCode  string  `json:"country_code"`
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ASN          uint    `json:"asn"`
	Organization string  `json:"organization"`
}
