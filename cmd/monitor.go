package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"traffic-mon/pkg/capture"
	"traffic-mon/pkg/geoip"
	"traffic-mon/pkg/metrics"
)

var (
	port        int
	metricsPort int
	device      string
	asnDB       string
	cityDB      string
	countryDB   string
	skipPrivate bool
	proxyProto  bool
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start monitoring traffic",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize GeoIP
		lookup, err := geoip.NewLookupService(asnDB, cityDB, countryDB)
		if err != nil {
			logger.Fatalf("Failed to initialize GeoIP: %v", err)
		}
		defer lookup.Close()

		logger.Infof("GeoIP databases loaded successfully")

		// Determine device
		if device == "" {
			device, err = capture.FindDevice()
			if err != nil {
				logger.Fatalf("Failed to find network device: %v", err)
			}
		}
		logger.Infof("Using network device: %s", device)

		// Start Metrics Server
		go func() {
			addr := fmt.Sprintf(":%d", metricsPort)
			logger.Infof("Starting metrics server on %s/metrics", addr)
			if err := metrics.StartMetricsServer(addr); err != nil {
				logger.Errorf("Metrics server failed: %v", err)
			}
		}()

		// Handle interrupts
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		// Start Capture
		go func() {
			err := capture.StartCapture(device, port, proxyProto, func(srcIP string) {
				if skipPrivate {
					ip := net.ParseIP(srcIP)
					if ip != nil && ip.IsPrivate() {
						return
					}
				}

				record, err := lookup.Lookup(srcIP)
				if err != nil {
					logger.Warnf("Lookup failed for %s: %v", srcIP, err)
					return
				}

				// Log the hit with colors
				logger.WithFields(logrus.Fields{
					"ip":           record.IP,
					"country":      record.Country,
					"country_code": record.CountryCode,
					"city":         record.City,
					"lat":          record.Latitude,
					"long":         record.Longitude,
					"asn":          record.ASN,
					"org":          record.Organization,
				}).Info("Traffic Hit")

				// Update Prometheus
				metrics.RecordHit(record)
			})
			if err != nil {
				logger.Fatalf("Capture failed: %v", err)
			}
		}()

		<-sigChan
		logger.Info("Shutting down...")
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().IntVarP(&port, "port", "p", 80, "Port to monitor")
	monitorCmd.Flags().IntVarP(&metricsPort, "metrics-port", "m", 9090, "Port for Prometheus metrics")
	monitorCmd.Flags().StringVarP(&device, "device", "d", "", "Network interface to use (default: auto-detect)")

	// Default paths based on user's reference
	monitorCmd.Flags().StringVar(&asnDB, "asn-db", "/home/hornet/personal-pr/cidr/GeoLite2-ASN.mmdb", "Path to ASN MMDB")
	monitorCmd.Flags().StringVar(&cityDB, "city-db", "/home/hornet/personal-pr/cidr/GeoLite2-City.mmdb", "Path to City MMDB")
	monitorCmd.Flags().StringVar(&countryDB, "country-db", "/home/hornet/personal-pr/cidr/GeoLite2-Country.mmdb", "Path to Country MMDB")
	monitorCmd.Flags().BoolVar(&skipPrivate, "skip-private", false, "Skip private/internal IP addresses")
	monitorCmd.Flags().BoolVar(&proxyProto, "proxy-protocol", false, "Enable PROXY protocol parsing")
}
