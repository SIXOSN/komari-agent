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
func NewRouteTraceTask(conn *ws.SafeConn, taskID uint, target string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	select {
	case routeTraceSlots <- struct{}{}:
		defer func() { <-routeTraceSlots }()
	case <-ctx.Done():
		return
	}

	var hops []string
	var traceErr error
	host, portText, err := net.SplitHostPort(target)
	if err != nil {
		host, portText = target, "80"
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 || strings.TrimSpace(host) == "" {
		traceErr = fmt.Errorf("invalid TCP Ping target")
	} else {
		hops, traceErr = traceTCPRoute(ctx, host, port)
	}

	errText := ""
	if traceErr != nil {
		errText = traceErr.Error()
		log.Printf("Route trace task %d failed: %s", taskID, errText)
	}
	payload := v2.BuildRouteResultPayload(taskID, target, hops, errText, time.Now().UTC())
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
