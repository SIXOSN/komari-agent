//go:build linux

package server

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
)

func traceTCPRouteIPv6(ctx context.Context, host string, port int) ([]string, error) {
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip6", host)
	if err != nil || len(addresses) == 0 {
		return nil, errors.New("no IPv6 target address")
	}
	target := addresses[0].To16()
	if target == nil || target.To4() != nil {
		return nil, errors.New("no IPv6 target address")
	}
	receiver, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
	if err != nil {
		return nil, fmt.Errorf("raw ICMPv6 socket unavailable: %w", err)
	}
	defer receiver.Close()

	hops := make([]string, 0, maxRouteHops)
	for ttl := 1; ttl <= maxRouteHops && ctx.Err() == nil; ttl++ {
		localPort := int(40000 + routeSourcePort.Add(1)%20000)
		probeCtx, cancel := context.WithTimeout(ctx, routeHopTimeout)
		result := make(chan routeDialResult, 1)
		go func(ttl, sourcePort int) {
			dialer := net.Dialer{
				LocalAddr: &net.TCPAddr{Port: sourcePort},
				Control: func(_, _ string, raw syscall.RawConn) error {
					var socketErr error
					if err := raw.Control(func(fd uintptr) {
						socketErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, ttl)
					}); err != nil {
						return err
					}
					return socketErr
				},
			}
			conn, dialErr := dialer.DialContext(probeCtx, "tcp6", net.JoinHostPort(target.String(), strconv.Itoa(port)))
			if conn != nil {
				conn.Close()
			}
			result <- routeDialResult{reached: dialErr == nil || errors.Is(dialErr, syscall.ECONNREFUSED)}
		}(ttl, localPort)

		hop := ""
		reached := false
		for probeCtx.Err() == nil {
			select {
			case dialResult := <-result:
				reached = dialResult.reached
			default:
			}
			if reached {
				hop = target.String()
				break
			}
			deadline := time.Now().Add(120 * time.Millisecond)
			if probeDeadline, ok := probeCtx.Deadline(); ok && deadline.After(probeDeadline) {
				deadline = probeDeadline
			}
			_ = receiver.SetReadDeadline(deadline)
			packet := make([]byte, 1500)
			n, peer, readErr := receiver.ReadFrom(packet)
			if readErr != nil {
				continue
			}
			message, parseErr := icmp.ParseMessage(syscall.IPPROTO_ICMPV6, packet[:n])
			if parseErr != nil {
				continue
			}
			var quoted []byte
			switch body := message.Body.(type) {
			case *icmp.TimeExceeded:
				quoted = body.Data
			case *icmp.DstUnreach:
				quoted = body.Data
			default:
				continue
			}
			if !routeQuotedTCPv6Matches(quoted, target, localPort, port) {
				continue
			}
			if ipPeer, ok := peer.(*net.IPAddr); ok {
				hop = ipPeer.IP.String()
			}
			if _, ok := message.Body.(*icmp.DstUnreach); ok {
				reached = true
			}
			break
		}
		cancel()
		hops = append(hops, hop)
		if reached || hop == target.String() {
			return hops, nil
		}
	}
	if ctx.Err() != nil {
		return hops, ctx.Err()
	}
	return hops, nil
}

// ICMPv6 errors quote the original 40-byte IPv6 header followed by TCP ports.
func routeQuotedTCPv6Matches(data []byte, target net.IP, sourcePort, destPort int) bool {
	if len(data) < 44 || data[0]>>4 != 6 || data[6] != syscall.IPPROTO_TCP || !net.IP(data[24:40]).Equal(target) {
		return false
	}
	return int(binary.BigEndian.Uint16(data[40:42])) == sourcePort &&
		int(binary.BigEndian.Uint16(data[42:44])) == destPort
}
