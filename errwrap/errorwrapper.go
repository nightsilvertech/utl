package errwrap

import (
	"errors"
	"fmt"
)

func Wrap(funcName, whenExec string, err error) error {
	message := fmt.Sprintf(
		"error happend on %s function, when execute %s function, detail error %+v",
		funcName,
		whenExec,
		err,
	)
	return errors.New(message)
}
