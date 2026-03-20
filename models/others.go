package models

type Ports struct {
	FiberServer string `json:"fiberServer"`
}

type LimiterResponse struct {
	Allowed       bool
	RetryAfter    int64
	CurrentTokens int64
}
