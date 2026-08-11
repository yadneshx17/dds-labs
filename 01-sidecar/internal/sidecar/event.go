package sidecar

type LogEvent struct {
	Service    string `json:"service"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	StatusCode int    `json:"statuscode"`
	WorkerId   int    `json:"worker_id"`
}
