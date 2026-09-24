//go:build linux

package server

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	maxRouteHops = 28
	routeHopTimeout = 900 * time.Millisecond
)

var routeSourcePort atomic.Uint32

type routeDialResult struct {
	reached bool
}

func traceTCPRoute(ctx context.Context, host string, port int) ([]string, error) {
	addresses, err := net.DefaultResolver.LookupIP(ctx, "ip4", host)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve IPv4 target: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("no IPv4 target address")
	}
	target := addresses[0].To4()
	if target == nil {
		return nil, fmt.Errorf("IPv4 target required")
	}

	receiver, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("raw ICMP socket unavailable: %w", err)
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
						socketErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, ttl)
					}); err != nil {
						return err
					}
					return socketErr
				},
			}
			conn, dialErr := dialer.DialContext(probeCtx, "tcp4", net.JoinHostPort(target.String(), strconv.Itoa(port)))
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
			message, parseErr := icmp.ParseMessage(1, packet[:n])
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
			if !routeQuotedTCPMatches(quoted, target, localPort, port) {
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

func routeQuotedTCPMatches(data []byte, target net.IP, sourcePort, destPort int) bool {
	header, err := ipv4.ParseHeader(data)
	if err != nil || header.Protocol != syscall.IPPROTO_TCP || !header.Dst.Equal(target) {
		return false
	}
	if len(data) < header.Len+4 {
		return false
	}
	return int(binary.BigEndian.Uint16(data[header.Len:])) == sourcePort &&
		int(binary.BigEndian.Uint16(data[header.Len+2:])) == destPort
}
