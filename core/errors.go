package core

// Error is an interface which aggregates the interface of type error and adds
// an additional method for a formatted error message.
type Error interface {
	error            // Reports error string in context-friendly format
	GetText() string // Reports error in a human-readable format
}
