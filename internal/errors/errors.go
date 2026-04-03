package errors

import "fmt"

type Friendly struct {
	Message string
	Cause   error
}

func (e Friendly) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return fmt.Sprintf("%s: %v", e.Message, e.Cause)
}

func Wrap(message string, err error) error {
	return Friendly{Message: message, Cause: err}
}
