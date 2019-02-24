package manager

import (
	"errors"
	"log"
	"testing"
)

func TestNewError(t *testing.T) {
	err := NewError("test", errors.New("test"), TestNewError, "1", "2")
	errMsg := err.Error()
	if errMsg != "test,test,proxy-golang/manager.TestNewError(1, 2)" {
		t.FailNow()
	}
	log.Printf("%v", errMsg)
}
