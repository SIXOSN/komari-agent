//go:build linux

package server

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestRouteQuotedTCPv6Matches(t *testing.T) {
	target := net.ParseIP("2001:db8::2")
	packet := make([]byte, 44)
	packet[0] = 0x60
	packet[6] = 6
	copy(packet[24:40], target.To16())
	binary.BigEndian.PutUint16(packet[40:42], 41000)
	binary.BigEndian.PutUint16(packet[42:44], 443)
	if !routeQuotedTCPv6Matches(packet, target, 41000, 443) {
		t.Fatal("valid quoted IPv6 TCP packet did not match")
	}
	if routeQuotedTCPv6Matches(packet, target, 41001, 443) {
		t.Fatal("wrong source port matched")
	}
	if routeQuotedTCPv6Matches(packet[:39], target, 41000, 443) {
		t.Fatal("truncated packet matched")
	}
}
