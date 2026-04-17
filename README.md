# Traffic-Mon

A real-time, high-performance network traffic monitor written in Go. It captures packets on any port, enriches them with local GeoIP data (ASN, City, Country), and exports analytics as Prometheus metrics for Grafana visualization.

## Features

- **Real-time Capture:** Direct packet capture using `gopacket` (libpcap).
- **Local GeoIP:** Blazing fast lookups using local MaxMind `.mmdb` files (no external API calls during capture).
- **Traffic Filtering:** Optionally skip internal/private IP ranges (RFC 1918) to focus on external traffic.
- **Proxy Support:** Supports **PROXY Protocol v1** to see real visitor IPs even when behind a load balancer.
- **Observability:**
  - **Colored CLI:** Real-time log of traffic hits with detailed location data.
  - **Prometheus Metrics:** Exported hit counters segmented by IP, Country, City, and ASN.
- **Configurable:** Monitor any port and customize metrics endpoints via CLI flags.

## Prerequisites

- **libpcap:** The system must have `libpcap` installed.
  - Ubuntu/Debian: `sudo apt-get install libpcap-dev`
- **GeoIP Databases:** Requires `.mmdb` files (ASN, City, Country). By default, it looks in `/home/hornet/personal-pr/cidr/`.
- **Root Privileges:** Live packet capture requires `sudo`.

## Installation

1. Clone the repository and navigate to the project root.
2. Build the binary:
   ```bash
   go build -o traffic-mon
   ```

## Usage

### Basic Monitoring
Monitor traffic on port 80 (default):
```bash
sudo ./traffic-mon monitor
```

### Custom Port, Metrics and Filtering
Monitor port 443, serve metrics on port 9100, skip internal traffic, and enable PROXY protocol:
```bash
sudo ./traffic-mon monitor --port 443 --metrics-port 9100 --skip-private --proxy-protocol
```

### Advanced Configuration
If your GeoIP files are in a different location:
```bash
sudo ./traffic-mon monitor \
  --asn-db ./db/ASN.mmdb \
  --city-db ./db/City.mmdb \
  --country-db ./db/Country.mmdb
```

## Analytics & Visualization

### 1. Terminal Output
The monitor displays a live-colored stream of incoming traffic:
```text
INFO[0005] Traffic Hit  asn=15169 city="Mountain View" country=USA ip=8.8.8.8 org="Google LLC"
```

### 2. Prometheus Metrics
Metrics are available at `http://localhost:9090/metrics`.
- **Metric Name:** `traffic_monitor_hits_total`
- **Labels:** `ip`, `country`, `city`, `asn`, `organization`

### 3. Grafana Integration
1. Add a **Prometheus Data Source** pointing to your monitor's metrics port.
2. Create dashboards using PromQL:
   - **Total Traffic by Country:** `sum(traffic_monitor_hits_total) by (country)`
   - **Top 10 Source IPs:** `topk(10, sum(traffic_monitor_hits_total) by (ip))`
   - **Traffic by ASN:** `sum(traffic_monitor_hits_total) by (asn)`

## Architecture

- `pkg/capture`: Manages raw socket capture and BPF filtering.
- `pkg/geoip`: Local MMDB lookup engine.
- `pkg/metrics`: Prometheus registry and hit recording.
- `cmd/`: CLI command structure using Cobra.

## Troubleshooting

### Seeing Private IPs (e.g., 10.x.x.x)
If you see private IPs in your logs even when external visitors are hitting your site, it typically means your server is behind a **Load Balancer** or **Reverse Proxy**. The monitor is seeing the internal IP of the proxy.
- Use `--skip-private` to filter these out.
- Use `--proxy-protocol` if your Load Balancer (e.g., GCP, AWS, Nginx) is configured to send PROXY protocol headers. This will allow the monitor to extract the real client IP.
