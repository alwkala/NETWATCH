package netenv

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"time"
)

var defaultSTUNServers = []string{
	"stun.cloudflare.com:3478",
	"stun.l.google.com:19302",
	"stun1.l.google.com:19302",
}

// DiscoverPublicIP queries lightweight STUN (RFC 5389 NAT Traversal)
// to resolve the external public IPv4 address without external HTTP egress.
func DiscoverPublicIP(ctx context.Context) string {
	for _, s := range defaultSTUNServers {
		select {
		case <-ctx.Done():
			return ""
		default:
		}
		ip, err := querySTUN(ctx, s)
		if err == nil && ip != "" {
			return ip
		}
	}
	return ""
}

// querySTUN sends a standard RFC 5389 Binding Request over raw UDP.
func querySTUN(ctx context.Context, server string) (string, error) {
	d := net.Dialer{Timeout: 1500 * time.Millisecond}
	conn, err := d.DialContext(ctx, "udp", server)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	var req [20]byte
	binary.BigEndian.PutUint16(req[0:2], 0x0001)     // Binding Request
	binary.BigEndian.PutUint16(req[2:4], 0x0000)     // Message Length: 0
	binary.BigEndian.PutUint32(req[4:8], 0x2112A442) // Magic Cookie
	if _, err := rand.Read(req[8:20]); err != nil {
		return "", err
	}

	_ = conn.SetDeadline(time.Now().Add(1500 * time.Millisecond))
	if _, err := conn.Write(req[:]); err != nil {
		return "", err
	}

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
	if n < 20 {
		return "", fmt.Errorf("stun: response too short")
	}

	msgType := binary.BigEndian.Uint16(buf[0:2])
	if msgType != 0x0101 { // Binding Success Response
		return "", fmt.Errorf("stun: not a success response (%x)", msgType)
	}

	pos := 20
	for pos+4 <= n {
		attrType := binary.BigEndian.Uint16(buf[pos : pos+2])
		attrLen := int(binary.BigEndian.Uint16(buf[pos+2 : pos+4]))
		pos += 4
		if pos+attrLen > n {
			break
		}

		// 0x0020 = XOR-MAPPED-ADDRESS
		if attrType == 0x0020 && attrLen >= 8 {
			family := buf[pos+1]
			if family == 0x01 { // IPv4
				xip := binary.BigEndian.Uint32(buf[pos+4 : pos+8])
				ip := xip ^ 0x2112A442
				addr := netip.AddrFrom4([4]byte{byte(ip >> 24), byte(ip >> 16), byte(ip >> 8), byte(ip)})
				if addr.IsValid() && addr.Is4() && !addr.IsPrivate() && !addr.IsLoopback() {
					return addr.String(), nil
				}
			}
		} else if attrType == 0x0001 && attrLen >= 8 { // MAPPED-ADDRESS (RFC 3489)
			family := buf[pos+1]
			if family == 0x01 {
				addr := netip.AddrFrom4([4]byte{buf[pos+4], buf[pos+5], buf[pos+6], buf[pos+7]})
				if addr.IsValid() && addr.Is4() && !addr.IsPrivate() && !addr.IsLoopback() {
					return addr.String(), nil
				}
			}
		}
		pos += (attrLen + 3) &^ 3
	}

	return "", fmt.Errorf("stun: no mapped address")
}
