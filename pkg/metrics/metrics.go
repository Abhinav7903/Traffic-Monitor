package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"traffic-mon/pkg/models"
)

var (
	TrafficHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_monitor_hits_total",
			Help: "Total number of traffic hits monitored",
		},
		[]string{"ip", "country", "country_code", "region", "city", "latitude", "longitude", "postal_code", "asn", "organization"},
	)
)

func init() {
	prometheus.MustRegister(TrafficHits)
}

func RecordHit(record *models.GeoRecord) {
	TrafficHits.WithLabelValues(
		record.IP,
		record.Country,
		record.CountryCode,
		record.Region,
		record.City,
		fmt.Sprintf("%.4f", record.Latitude),
		fmt.Sprintf("%.4f", record.Longitude),
		record.PostalCode,
		fmt.Sprintf("%d", record.ASN),
		record.Organization,
	).Inc()
}

// Fixed RecordHit below in a later thought if I notice the bug, or I can fix it now.
// Actually, string(rune(record.ASN)) is definitely wrong.

func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}
