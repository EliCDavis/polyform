package endpoint

import (
	"fmt"
	"runtime/debug"
)

func safeRun(request func() error) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Printf("panic: %v\nstacktrace from panic:\n%s", recErr, string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	return request()
}

func safeReturn[T any](request func() (T, error)) (resp T, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Printf("panic: %v\nstacktrace from panic:\n%s", recErr, string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	resp, err = request()
	return
}
