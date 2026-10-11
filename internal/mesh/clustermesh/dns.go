package clustermesh

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

// DNSInterceptor handles split-horizon DNS queries for the global service domain
type DNSInterceptor struct {
	mgr      *Manager
	port     int
	listener *net.UDPConn
}

// NewDNSInterceptor creates a DNSInterceptor bound to a ClusterMesh Manager
func NewDNSInterceptor(mgr *Manager, port int) *DNSInterceptor {
	if port == 0 {
		port = 5353
	}
	return &DNSInterceptor{
		mgr:  mgr,
		port: port,
	}
}

// Start launches the UDP DNS listener in the background
func (d *DNSInterceptor) Start(ctx context.Context) error {
	addr := &net.UDPAddr{
		Port: d.port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind UDP DNS interceptor on :%d: %w", d.port, err)
	}
	d.listener = conn

	go func() {
		defer func() { _ = conn.Close() }()
		buf := make([]byte, 1024)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, remoteAddr, err := conn.ReadFrom(buf)
			if err != nil {
				continue
			}

			resp, err := d.ProcessDNSQuery(buf[:n])
			if err == nil && len(resp) > 0 {
				_, _ = conn.WriteTo(resp, remoteAddr)
			}
		}
	}()

	return nil
}

// Close stops the DNS interceptor
func (d *DNSInterceptor) Close() error {
	if d.listener != nil {
		return d.listener.Close()
	}
	return nil
}

// ProcessDNSQuery parses a raw DNS request, resolves global service names, and generates the wire response
func (d *DNSInterceptor) ProcessDNSQuery(req []byte) ([]byte, error) {
	if len(req) < 12 {
		return nil, fmt.Errorf("dns query packet too short (%d bytes)", len(req))
	}

	txID := binary.BigEndian.Uint16(req[0:2])
	qdCount := binary.BigEndian.Uint16(req[4:6])

	if qdCount == 0 {
		return nil, fmt.Errorf("no questions in dns packet")
	}

	// Parse Question QNAME
	offset := 12
	qname, nextOffset, err := parseDomainName(req, offset)
	if err != nil {
		return nil, fmt.Errorf("malformed qname: %w", err)
	}

	if len(req) < nextOffset+4 {
		return nil, fmt.Errorf("dns query truncated in question section")
	}

	qtype := binary.BigEndian.Uint16(req[nextOffset : nextOffset+2])
	qclass := binary.BigEndian.Uint16(req[nextOffset+2 : nextOffset+4])

	qnameLower := strings.ToLower(qname)
	suffix := strings.ToLower(d.mgr.cfg.DomainSuffix)

	// If domain does not match cluster global suffix, return NXDOMAIN
	if !strings.HasSuffix(qnameLower, suffix) {
		return buildNXDomainResponse(txID, req[12:nextOffset+4]), nil
	}

	vip, found := d.mgr.ResolveGlobalVIP(qnameLower)
	if !found || vip == nil {
		return buildNXDomainResponse(txID, req[12:nextOffset+4]), nil
	}

	// Only synthesize A records for IPv4 VIPs
	if qtype == 1 /* A */ || qtype == 255 /* ANY */ {
		return buildAResponse(txID, req[12:nextOffset+4], qname, qclass, vip)
	}

	// For AAAA or others, return empty NoError (NODATA)
	return buildNoDataResponse(txID, req[12:nextOffset+4]), nil
}

func parseDomainName(packet []byte, start int) (string, int, error) {
	var parts []string
	curr := start

	for {
		if curr >= len(packet) {
			return "", 0, fmt.Errorf("offset out of bounds")
		}
		labelLen := int(packet[curr])
		if labelLen == 0 {
			curr++
			break
		}
		curr++
		if curr+labelLen > len(packet) {
			return "", 0, fmt.Errorf("label exceeds packet boundary")
		}
		parts = append(parts, string(packet[curr:curr+labelLen]))
		curr += labelLen
	}

	return strings.Join(parts, "."), curr, nil
}

func encodeDomainName(domain string) []byte {
	var buf bytes.Buffer
	labels := strings.Split(domain, ".")
	for _, label := range labels {
		if len(label) == 0 {
			continue
		}
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0)
	return buf.Bytes()
}

func buildAResponse(txID uint16, questionBytes []byte, _ string, qclass uint16, ip net.IP) ([]byte, error) {
	var buf bytes.Buffer

	// Header: 12 bytes
	_ = binary.Write(&buf, binary.BigEndian, txID)
	// Flags: QR=1 (response), AA=1 (authoritative), RD=1, RA=1, RCODE=0
	_ = binary.Write(&buf, binary.BigEndian, uint16(0x8580))
	_ = binary.Write(&buf, binary.BigEndian, uint16(1)) // QDCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(1)) // ANCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // NSCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // ARCount

	// Echo Question
	buf.Write(questionBytes)

	// Answer: Pointer to question name (0xc00c)
	_ = binary.Write(&buf, binary.BigEndian, uint16(0xc00c))
	_ = binary.Write(&buf, binary.BigEndian, uint16(1))  // Type: A
	_ = binary.Write(&buf, binary.BigEndian, qclass)     // Class: IN (1)
	_ = binary.Write(&buf, binary.BigEndian, uint32(30)) // TTL: 30s
	_ = binary.Write(&buf, binary.BigEndian, uint16(4))  // RDLength: 4 bytes

	ipv4 := ip.To4()
	if ipv4 == nil {
		ipv4 = net.IPv4(0, 0, 0, 0).To4()
	}
	buf.Write(ipv4)

	return buf.Bytes(), nil
}

func buildNXDomainResponse(txID uint16, questionBytes []byte) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, txID)
	// Flags: QR=1, AA=1, RCODE=3 (NXDOMAIN) -> 0x8583
	_ = binary.Write(&buf, binary.BigEndian, uint16(0x8583))
	_ = binary.Write(&buf, binary.BigEndian, uint16(1)) // QDCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // ANCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // NSCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // ARCount
	buf.Write(questionBytes)
	return buf.Bytes()
}

func buildNoDataResponse(txID uint16, questionBytes []byte) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, txID)
	// Flags: QR=1, AA=1, RCODE=0 -> 0x8580
	_ = binary.Write(&buf, binary.BigEndian, uint16(0x8580))
	_ = binary.Write(&buf, binary.BigEndian, uint16(1)) // QDCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // ANCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // NSCount
	_ = binary.Write(&buf, binary.BigEndian, uint16(0)) // ARCount
	buf.Write(questionBytes)
	return buf.Bytes()
}
