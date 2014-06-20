package error

import (
	"errors"
)

var (
	UnknownQuery = NewErr("unknown bson.M query")
	ErrNotFound  = NewErr("not found")
	ErrTypeAll   = NewErr("result argument must be a slice address")
	ErrType      = NewErr("result argument must be a address")
)

func NewErr(err string) error {
	return errors.New(err)
}
