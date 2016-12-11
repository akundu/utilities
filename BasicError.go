package utilities

// errorString is a trivial implementation of error.
type BasicError struct {
	s string
}

func (e *BasicError) Error() string {
	return e.s
}

func NewBasicError(s string) *BasicError {
	return &BasicError{s}
}
