//go:build linux

package server

import (
	"context"
	"encoding/binary"
	"net"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestRouteQuotedTCPMatches(t *testing.T) {
	packet := make([]byte, 28)
	packet[0] = 0x45
	binary.BigEndian.PutUint16(packet[2:4], uint16(len(packet)))
	packet[8] = 64
	packet[9] = 6 // TCP
	copy(packet[12:16], net.ParseIP("192.0.2.10").To4())
	copy(packet[16:20], net.ParseIP("202.96.128.86").To4())
	binary.BigEndian.PutUint16(packet[20:22], 45000)
	binary.BigEndian.PutUint16(packet[22:24], 443)
	target := net.ParseIP("202.96.128.86")
	if !routeQuotedTCPMatches(packet, target, 45000, 443) {
		t.Fatal("matching quoted TCP probe was rejected")
	}
	if routeQuotedTCPMatches(packet, target, 45001, 443) || routeQuotedTCPMatches(packet, target, 45000, 80) {
		t.Fatal("unrelated quoted TCP probe was accepted")
	}
}

func TestTraceTCPRouteLive(t *testing.T) {
	if os.Getenv("KOMARI_ROUTE_TRACE_LIVE") != "1" {
		t.Skip("requires an IPv4 route and CAP_NET_RAW")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	target := os.Getenv("KOMARI_ROUTE_TRACE_TARGET")
	if target == "" {
		target = "1.1.1.1:443"
	}
	host, portText, err := net.SplitHostPort(target)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	hops, err := traceTCPRoute(ctx, host, port)
	if err != nil {
		t.Fatal(err)
	}
	for _, hop := range hops {
		if hop != "" {
			return
		}
	}
	t.Fatalf("trace returned no responding hops: %v", hops)
}
