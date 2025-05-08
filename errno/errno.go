package errno

import (
	"fmt"
)

// Errno ...
type Errno struct {
	Code    int
	Message string
}

func (err Errno) Error() string {
	return err.Message
}

// Err represents an error
type Err struct {
	Code    int
	Message string
	Err     string
}

// New ...
func New(errno *Errno, err error) *Err {
	return &Err{Code: errno.Code, Message: err.Error(), Err: err.Error()}
}

// Add ...
func (err *Err) Add(message string) error {
	err.Message += " " + message
	return err
}

// AddFormat ...
func (err *Err) AddFormat(format string, args ...interface{}) error {
	err.Message += " " + fmt.Sprintf(format, args...)
	return err
}

// Error ...
func (err *Err) Error() string {
	return fmt.Sprintf("Err - code: %d, message: %s, error: %s", err.Code, err.Message, err.Err)
}

// DecodeErr ...
func DecodeErr(err error) (int, bool, string) {
	if err == nil {
		return OK.Code, true, OK.Message
	}

	switch typed := err.(type) {
	case *Err:
		return typed.Code, false, typed.Message
	case *Errno:
		return typed.Code, false, typed.Message
	default:
	}

	return InternalServerError.Code, false, err.Error()
}
