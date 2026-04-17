package capture

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketHandler func(srcIP string)

func StartCapture(device string, port int, proxyProto bool, handler PacketHandler) error {
	handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("error opening device %s: %v", device, err)
	}
	defer handle.Close()

	// Filter for TCP traffic on the target port (broad filter for reliability)
	filter := fmt.Sprintf("tcp port %d", port)
	if err := handle.SetBPFFilter(filter); err != nil {
		return fmt.Errorf("error setting BPF filter: %v", err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	log.Printf("Starting capture on device %s, port %d...", device, port)

	for packet := range packetSource.Packets() {
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		if tcpLayer == nil {
			continue
		}
		tcp, _ := tcpLayer.(*layers.TCP)

		// Only process incoming packets to our monitored port
		if int(tcp.DstPort) != port {
			continue
		}

		// Log if it's a new connection (SYN) OR if it contains data (PSH flag with payload)
		// This helps capture multiple requests on the same Keep-Alive connection
		isNewConn := tcp.SYN && !tcp.ACK
		isData := tcp.PSH && len(tcp.Payload) > 0

		if !isNewConn && !isData {
			continue
		}

		srcIP := ""
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			srcIP = ip.SrcIP.String()
		} else {
			ipv6Layer := packet.Layer(layers.LayerTypeIPv6)
			if ipv6Layer != nil {
				ip, _ := ipv6Layer.(*layers.IPv6)
				srcIP = ip.SrcIP.String()
			}
		}

		if srcIP == "" {
			continue
		}

		// If PROXY protocol is enabled, try to extract the real client IP from the payload
		if proxyProto && len(tcp.Payload) > 0 {
			payloadStr := string(tcp.Payload)
			if strings.HasPrefix(payloadStr, "PROXY ") {
				// PROXY TCP4/TCP6 SOURCE_IP DEST_IP SOURCE_PORT DEST_PORT\r\n
				parts := strings.Split(payloadStr, " ")
				if len(parts) >= 3 {
					srcIP = parts[2]
				}
			}
		}

		handler(srcIP)
	}

	return nil
}

func FindDevice() (string, error) {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", fmt.Errorf("no devices found")
	}
	// Return first device with an IP as a heuristic
	for _, d := range devices {
		if len(d.Addresses) > 0 {
			return d.Name, nil
		}
	}
	return devices[0].Name, nil
}
