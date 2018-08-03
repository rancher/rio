package client

const (
	RetryType               = "retry"
	RetryFieldAttempts      = "attempts"
	RetryFieldTimeoutMillis = "timeoutMillis"
)

type Retry struct {
	Attempts      int64 `json:"attempts,omitempty" yaml:"attempts,omitempty"`
	TimeoutMillis int64 `json:"timeoutMillis,omitempty" yaml:"timeoutMillis,omitempty"`
}
