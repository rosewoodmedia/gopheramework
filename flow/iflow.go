package flow

import "github.com/rosewoodmedia/gopheramework/core"

type IFlow interface {
	core.Error

	// Error-reporting methods

	Add(err error) bool
	Check(err error, label string) bool

	// Assertion methods

	Must(check bool, message string) bool

	// Debugging/logging methods

	Log(note string)

	// Flow terminatino methods

	Abort(status string) error
	Done(status string) error
}
