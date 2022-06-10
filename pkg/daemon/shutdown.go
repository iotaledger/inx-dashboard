package daemon

const (
	PriorityDisconnectINX = iota // no dependencies
	PriorityStopDashboard
	PriorityStopPrometheus
)
