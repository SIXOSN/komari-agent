package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	v2 "github.com/komari-monitor/komari-agent/protocol/v2"
	"github.com/komari-monitor/komari-agent/ws"
)

var routeTraceSlots = make(chan struct{}, 1)

// NewRouteTraceTask probes the path to the same host and port as a TCP Ping task.
// Only one trace runs at a time on an agent to bound raw socket and network use.
func NewRouteTraceTask(conn *ws.SafeConn, taskID uint, target, family string) {
	if family == "" {
		family = "ipv4"
	}
	if family != "ipv4" && family != "ipv6" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	select {
	case routeTraceSlots <- struct{}{}:
		defer func() { <-routeTraceSlots }()
	case <-ctx.Done():
		return
	}

	var hops []string
	var samples [][]string
	var traceErr error
	resolvedIP := ""
	attempts := 0
	host, portText, err := net.SplitHostPort(target)
	if err != nil {
		host, portText = target, "80"
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 || strings.TrimSpace(host) == "" {
		traceErr = fmt.Errorf("invalid TCP Ping target")
	} else {
		addresses, lookupErr := net.DefaultResolver.LookupIP(ctx, map[string]string{"ipv4": "ip4", "ipv6": "ip6"}[family], host)
		if lookupErr != nil || len(addresses) == 0 {
			if family == "ipv6" {
				traceErr = fmt.Errorf("no IPv6 target address")
			} else {
				traceErr = fmt.Errorf("cannot resolve IPv4 route target")
			}
		} else {
			resolvedIP = addresses[0].String()
			hops, traceErr = traceTCPRoute(ctx, resolvedIP, port, family)
			attempts = 1
			samples = append(samples, append([]string(nil), hops...))
			// A single silent router can hide the mainland entry. Retry only
			// incomplete paths, never more than three passes per scheduled task.
			// Do not splice different passes: route changes would create a path
			// that never actually existed.
			for attempts < 3 && traceErr == nil && routeHasMissingHop(hops) && ctx.Err() == nil {
				more, retryErr := traceTCPRoute(ctx, resolvedIP, port, family)
				attempts++
				samples = append(samples, append([]string(nil), more...))
				if retryErr != nil {
					break // Keep the usable first pass.
				}
				hops = bestRouteHops(hops, more)
			}
		}
	}

	errText := ""
	if traceErr != nil {
		errText = traceErr.Error()
		log.Printf("Route trace task %d failed: %s", taskID, errText)
	}
	payload := v2.BuildRouteResultPayloadWithSamples(taskID, target, family, resolvedIP, attempts, hops, samples, errText, time.Now().UTC())
	if conn == nil {
		if err := postV2RPC(payload); err != nil {
			log.Printf("Failed to upload route trace over POST: %v", err)
		}
		return
	}
	if err := conn.WriteJSON(payload); err != nil {
		log.Printf("Failed to upload route trace over WebSocket: %v", err)
	}
}

func routeHasMissingHop(hops []string) bool {
	if len(hops) == 0 {
		return true
	}
	for _, hop := range hops {
		if hop == "" {
			return true
		}
	}
	return false
}

func bestRouteHops(primary, retry []string) []string {
	answered := func(hops []string) int {
		count := 0
		for _, hop := range hops {
			if hop != "" {
				count++
			}
		}
		return count
	}
	if answered(retry) > answered(primary) {
		return retry
	}
	return primary
}
