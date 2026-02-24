package screen

// State represents the current screen state.
type State int

const (
	StateUnknown State = iota // Cannot determine screen state.
	StateActive               // Screen is on and unlocked.
	StateIdle                 // Screen is off or locked.
)
