package models

type GeoRecord struct {
	IP           string `json:"ip"`
	Country      string `json:"country"`
	City         string `json:"city"`
	ASN          uint   `json:"asn"`
	Organization string `json:"organization"`
}
