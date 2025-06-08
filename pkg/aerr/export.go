package aerr

import (
	"errors"
	"fmt"
)

var (
	New    = errors.New
	Errorf = fmt.Errorf
	Newf   = fmt.Errorf
	Is     = errors.Is
	Join   = errors.Join
	Unwrap = errors.Unwrap
	As     = errors.As

	// common errors
	ErrUnsupported = errors.ErrUnsupported
	ErrNilArg      = New("nil argument")
	ErrIllegalType = New("illegal type")
)
