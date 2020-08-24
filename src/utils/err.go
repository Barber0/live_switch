package utils

import (
	"errors"
	"fmt"
)

func CheckErr(err error, callbacks ...func(err error)) {
	if err != nil {
		for _, fn := range callbacks {
			fn(err)
		}
	}
}

func HandlePanic(fns ...func(err error)) {
	if pa := recover(); pa != nil {
		var err error
		switch pa.(type) {
		case error:
			err = pa.(error)
		case string:
			err = errors.New(pa.(string))
		default:
			err = fmt.Errorf("%s", pa)
		}
		for _, fn := range fns {
			fn(err)
		}
	}
}
