package flow_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rosewoodmedia/rwgo/lib/common/RC2018/q1"
)

func fakeLoginCheck(inputErr bool, loginErr bool, systemErr bool) error {
	f1 := q1.NewFlow("login check")
	f1.LogWithTime("login from 192.169.111.222")
	if inputErr {
		f2 := f1.Flow("bad input")
		return f2.Done("bad input")
	}
	if systemErr {
		f2 := f1.Flow("database error")
		f2.Log("flow diverted from login check")
		err := errors.New("database: generic database error")
		if f2.Check(err, "get user record") {
			f2.Log("critical: maybe logs can trigger certain handlers")
			return f2.Abort("system")
		}
	}
	if loginErr {
		return f1.Done("bad login")
	}
	return f1.Done("login okay")
}

func TestAltFlow(t *testing.T) {
	err := fakeLoginCheck(true, false, false)
	if err != nil {
		t.Error(err)
	}
	err = fakeLoginCheck(false, true, false)
	if err != nil {
		t.Error(err)
	}
}
func TestExcFlow(t *testing.T) {
	err := fakeLoginCheck(false, false, true)
	fmt.Println(q1.LabelError(err, "test").(q1.Error).GetText())
}
func TestNorFlow(t *testing.T) {
	err := fakeLoginCheck(false, false, false)
	if err != nil {
		t.Error(err)
	}
}
