package models

type GeoRecord struct {
	IP           string `json:"ip"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	City         string `json:"city"`
	ASN          uint   `json:"asn"`
	Organization string `json:"organization"`
}
