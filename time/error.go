package time

// E is the internal error type
type E string

// Error implements errors interface
func (e E) Error() string {
	return string(e)
}

const (
	// ErrTimerIsNotRunning indicates that timer is not running
	ErrTimerIsNotRunning = E("timer is not running")
	// ErrTimerIsAlreadyRunning indicates that timer is already running
	ErrTimerIsAlreadyRunning = E("timer is already running")
)
