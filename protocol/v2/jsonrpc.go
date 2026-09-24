package v2

import (
	"encoding/json"
	"time"
)

const (
	DistributionHeader     = "X-SIXOSN-Komari-Distribution"
	ServerDistribution     = "SIXOSN/komari"
	AgentDistribution      = "SIXOSN/komari-agent"
	Version                = "2.0"
	MethodAgentReport      = "agent.report"
	MethodAgentBasicInfo   = "agent.basicInfo"
	MethodAgentPingResult  = "agent.pingResult"
	MethodAgentPing        = "agent.ping"
	MethodAgentRouteTrace  = "agent.routeTrace"
	MethodAgentRouteResult = "agent.routeResult"
	MethodAgentMessage     = "agent.message"
	MethodAgentEvent       = "agent.event"
	MethodAgentPull        = "agent.pull"
)

type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Event struct {
	ID        string      `json:"id"`
	Method    string      `json:"method"`
	Params    interface{} `json:"params,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
}

type EventResult struct {
	Status string  `json:"status,omitempty"`
	Events []Event `json:"events,omitempty"`
}

func NewNotification(method string, params interface{}) []byte {
	payload, _ := json.Marshal(Request{JSONRPC: Version, Method: method, Params: params})
	return payload
}

func NewRequest(id interface{}, method string, params interface{}) []byte {
	payload, _ := json.Marshal(Request{JSONRPC: Version, Method: method, Params: params, ID: id})
	return payload
}

func BuildReportPayload(report []byte) []byte {
	return NewNotification(MethodAgentReport, reportParams{Report: json.RawMessage(report)})
}

func BuildReportRequest(id interface{}, report []byte, ackEventIDs []string) []byte {
	return NewRequest(id, MethodAgentReport, reportParams{Report: json.RawMessage(report), AckEventIDs: ackEventIDs})
}

func BuildBasicInfoPayload(info map[string]interface{}) []byte {
	return NewNotification(MethodAgentBasicInfo, map[string]interface{}{"info": info})
}

type reportParams struct {
	Report      json.RawMessage `json:"report"`
	AckEventIDs []string        `json:"ack_event_ids,omitempty"`
}

func BuildPingResultPayload(taskID uint, pingType string, value int, finishedAt time.Time) interface{} {
	return Request{
		JSONRPC: Version,
		Method:  MethodAgentPingResult,
		Params: map[string]interface{}{
			"task_id":     taskID,
			"ping_type":   pingType,
			"value":       value,
			"finished_at": finishedAt.Format(time.RFC3339Nano),
		},
	}
}

func BuildRouteResultPayload(taskID uint, target, family string, hops []string, traceError string, finishedAt time.Time) interface{} {
	return BuildRouteResultPayloadDetailed(taskID, target, family, "", 0, hops, traceError, finishedAt)
}

func BuildRouteResultPayloadDetailed(taskID uint, target, family, resolvedIP string, attempts int, hops []string, traceError string, finishedAt time.Time) interface{} {
	return BuildRouteResultPayloadWithSamples(taskID, target, family, resolvedIP, attempts, hops, nil, traceError, finishedAt)
}

func BuildRouteResultPayloadWithSamples(taskID uint, target, family, resolvedIP string, attempts int, hops []string, samples [][]string, traceError string, finishedAt time.Time) interface{} {
	return Request{
		JSONRPC: Version,
		Method:  MethodAgentRouteResult,
		Params: map[string]interface{}{
			"task_id":     taskID,
			"target":      target,
			"family":      family,
			"resolved_ip": resolvedIP,
			"attempts":    attempts,
			"hops":        hops,
			"samples":     samples,
			"error":       traceError,
			"finished_at": finishedAt.Format(time.RFC3339Nano),
		},
	}
}

func BindParams(raw interface{}, target interface{}) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, target)
}

func BindResult(raw interface{}, target interface{}) error {
	return BindParams(raw, target)
}
