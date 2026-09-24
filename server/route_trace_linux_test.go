//go:build linux

package server

import (
	"encoding/binary"
	"net"
	"testing"
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
