package manager

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

var (
	KeyNotExist = errors.New("ProxyRelayNotExist")
	KeyExist    = errors.New("ProxyRelayExist")
)

type OpError struct {
	// Op is the operation which caused the error, such as
	// "read" or "write".
	Op string
	// Err is the error that occurred during the operation.
	Err error
	// Method
	Method interface{}
	// Params
	Params []reflect.Value
}

func (self OpError) Error() (errDesc string) {
	MethodName := runtime.FuncForPC(reflect.ValueOf(self.Method).Pointer()).Name()
	var params string
	for _, param := range self.Params[0 : len(self.Params)-1] {
		params += fmt.Sprintf("%v, ", param)
	}
	params += fmt.Sprintf("%v", self.Params[len(self.Params)-1])
	errDesc = fmt.Sprintf("%v,%v,%v(%s)", self.Op, self.Err.Error(), MethodName, params)
	return
}
func NewError(Op string, Err error, Method interface{}, Params ...interface{}) OpError {
	var params []reflect.Value
	for _, value := range Params {
		params = append(params, reflect.ValueOf(value))
	}
	return OpError{
		Op:     Op,
		Err:    Err,
		Method: Method,
		Params: params,
	}
}
