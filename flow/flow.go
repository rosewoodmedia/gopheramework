package flow

import (
	"errors"
	"os"
	"runtime/debug"
	"strings"

	"github.com/rosewoodmedia/gopheramework/core"
	"github.com/rosewoodmedia/rwgo/lib/rwstrings"
)

// FlowOptions contains options for creating new flow instances. It can be
// considered a factory for Flow.
type FlowOptions struct {
	// FlowCheckWriteStack determines if stack traces arae written when fl.Check
	// is called.
	FlowCheckWriteStack bool
	// PrintStderrOnAbort determines if an error message gets printed to stderr
	// when Abort() is called.
	PrintStderrOnAbort bool
}

// NewDefault creates a flow that includes stack traces and does not output to
// stderr automatically.
func NewDefault(name string) IFlow {
	return FlowOptions{
		FlowCheckWriteStack: true,
		PrintStderrOnAbort:  false,
	}.New(name)
}

// NewDebug creates a flow that includes stack traces and will output to stderr
// whenever the Abort method is called.
func NewDebug(name string) IFlow {
	return FlowOptions{
		FlowCheckWriteStack: true,
		PrintStderrOnAbort:  false,
	}.New(name)
}

// New creates a Flow which aggregates the reciever (options) and is identified
// by the specified name. The name parameter does not need to be unique,
// although it is recommended to enforce an appropriate naming convention in a
// package where Flow is used.
func (options FlowOptions) New(name string) IFlow {
	return options.new(name)
}

func (options FlowOptions) new(name string) *Flow {
	return &Flow{
		options: options,
		Name:    name,

		children: []interface{}{},
		flows:    []*Flow{},
		logs:     []string{},
	}
}

// Flow represents the flow of a program for use in analytics and error
// reporting. A wrapper with additional application logic is generally the best
// way to use Flow (for example, log functions that prepend an error level)
type Flow struct {
	Name string

	options FlowOptions

	// children aggregates all the following types to keep the order in which
	// they were applied to the flow (for logging)
	// - error (error given by caller)
	// - Flow (passes type error) (created by method of this struct)
	// - Log (fails type error) (created by method of this struct)
	children []interface{}

	errors []error
	flows  []*Flow
	logs   []string
}

func (fl *Flow) add(i interface{}) {
	fl.children = append(fl.children, i)
}

func (fl *Flow) addError(err error) bool {
	if err == nil {
		return false
	}
	fl.errors = append(fl.errors, err)
	fl.add(err)
	return true
}

func (fl *Flow) addFlow(newFlow *Flow) {
	fl.flows = append(fl.flows, newFlow)
	fl.add(newFlow)
}

func (fl *Flow) addLog(note string) {
	fl.logs = append(fl.logs, note)
	fl.add(note)
}

func (fl *Flow) getError() error {
	if len(fl.errors) > 0 {
		return fl
	}
	return nil
}

// Add will, if and only if the passed error is not nil, add the error to the
// flow without wrapping it in a struct or adding any message.
func (fl *Flow) Add(err error) bool {
	return fl.addError(err)
}

// Check will, if and only if the passed error is not nil, wrap the error in
// an Error struct with the specified label and add it to the list of
// errors associated with the current segment.
func (fl *Flow) Check(err error, label string) bool {
	if err == nil {
		return false
	}
	erro := LabelledError{
		Err:   err,
		Label: label,
		Flow:  fl.Name,
	}
	if fl.options.FlowCheckWriteStack {
		erro.Trace = debug.Stack()
	}
	fl.addError(erro)
	return true
}

func (fl *Flow) Must(check bool, message string) bool {
	if !check {
		fl.Check(errors.New("flow.Must fail"), message)
		return true
	}
	return false
}

// Flow adds a flow as a child of the current flow. This acts as a named path
// that can be used for debugging and analytics. This method should be called on
// any significant conditional block - that is, any conditional code that could
// be considered an alternative flow in a use case diagram.
func (fl *Flow) Flow(name string) IFlow {
	newFlow := fl.options.new(name)
	newFlow.addFlow(fl)
	return newFlow
}

// Abort logs that the flow ended due to a system error. The flow is always
// reported as an error with the specified label.
func (fl *Flow) Abort(label string) error {
	erro := LabelledError{
		Err:   fl,
		Label: label,
		Flow:  fl.Name,
	}
	if fl.options.PrintStderrOnAbort {
		os.Stderr.WriteString(erro.GetText())
	}
	return erro
}

// Done logs that the flow was completed. The flow is reported as an error if
// there were any errors checked in the flow, otherwise nil is returned.
// Child flows are not checked for errors.
func (fl *Flow) Done(label string) error {
	err := fl.getError()
	if err != nil {
		err = LabelledError{
			Err:   err,
			Label: label,
			Flow:  fl.Name,
		}
	}
	return err
}

// Log adds a note to the flow that will appear in error strings
func (fl *Flow) Log(note string) {
	fl.addLog(note)
}

// Error reports a detailed context-friendly error string
func (fl *Flow) Error() string {
	errStrings := []string{}
	for _, err := range fl.errors {
		errStrings = append(errStrings, err.Error())
	}
	childStrings := []string{}
	for _, child := range fl.children {
		switch ch := child.(type) {
		// Go error
		case error:
			childStrings = append(childStrings, ch.Error())
		// Note
		case string:
			childStrings = append(childStrings, "note("+ch+")")
		}
	}
	return "flow(" + fl.Name +
		"): {" + strings.Join(childStrings, ";") + "}"
}

// GetText reports a detailed human-readable error string
func (fl *Flow) GetText() string {
	childStrings := []string{}
	for _, child := range fl.children {
		switch ch := child.(type) {
		// Flow or labelled error
		case core.Error:
			childStrings = append(childStrings,
				"\t- "+rwstrings.Indent(ch.GetText()),
			)
		// Go error
		case error:
			childStrings = append(childStrings,
				"\t- error: "+rwstrings.Indent(ch.Error())+"\n",
			)
		// Note
		case string:
			childStrings = append(childStrings, "\t- log: "+ch+"\n")
		}
	}
	if len(childStrings) == 0 {
		return "Flow " + fl.Name + ": (nothing to report)\n"
	}
	return "Flow " + fl.Name + ":\n" +
		strings.Join(childStrings, "")
}
