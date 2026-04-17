package geoip

import (
	"fmt"
	"net"

	"github.com/oschwald/maxminddb-golang"
	"traffic-mon/pkg/models"
)

type ASNRecord struct {
	AutonomousSystemNumber       uint   `maxminddb:"autonomous_system_number"`
	AutonomousSystemOrganization string `maxminddb:"autonomous_system_organization"`
}

type CityRecord struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Country struct {
		IsoCode string            `maxminddb:"iso_code"`
		Names   map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
}

type LookupService struct {
	dbASN     *maxminddb.Reader
	dbCity    *maxminddb.Reader
	dbCountry *maxminddb.Reader
}

func NewLookupService(asnPath, cityPath, countryPath string) (*LookupService, error) {
	dbASN, err := maxminddb.Open(asnPath)
	if err != nil {
		return nil, fmt.Errorf("error opening ASN database: %v", err)
	}

	dbCity, err := maxminddb.Open(cityPath)
	if err != nil {
		dbASN.Close()
		return nil, fmt.Errorf("error opening City database: %v", err)
	}

	dbCountry, err := maxminddb.Open(countryPath)
	if err != nil {
		dbASN.Close()
		dbCity.Close()
		return nil, fmt.Errorf("error opening Country database: %v", err)
	}

	return &LookupService{
		dbASN:     dbASN,
		dbCity:    dbCity,
		dbCountry: dbCountry,
	}, nil
}

func (s *LookupService) Close() {
	if s.dbASN != nil {
		s.dbASN.Close()
	}
	if s.dbCity != nil {
		s.dbCity.Close()
	}
	if s.dbCountry != nil {
		s.dbCountry.Close()
	}
}

func (s *LookupService) Lookup(ipStr string) (*models.GeoRecord, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", ipStr)
	}

	record := &models.GeoRecord{
		IP: ipStr,
	}

	var asnRec ASNRecord
	if err := s.dbASN.Lookup(ip, &asnRec); err == nil {
		record.ASN = asnRec.AutonomousSystemNumber
		record.Organization = asnRec.AutonomousSystemOrganization
	}

	var cityRec CityRecord
	if err := s.dbCity.Lookup(ip, &cityRec); err == nil {
		record.City = cityRec.City.Names["en"]
		record.Country = cityRec.Country.Names["en"]
		record.CountryCode = cityRec.Country.IsoCode
	}

	// Fallback to country DB if city DB didn't have country
	if record.Country == "" {
		var countryRec CityRecord
		if err := s.dbCountry.Lookup(ip, &countryRec); err == nil {
			record.Country = countryRec.Country.Names["en"]
			record.CountryCode = countryRec.Country.IsoCode
		}
	}

	return record, nil
}
