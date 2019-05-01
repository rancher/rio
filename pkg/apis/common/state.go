package common

type State struct {
	State         string
	Error         bool
	Transitioning bool
	Message       string
}

type StateGetter interface {
	State() State
}
