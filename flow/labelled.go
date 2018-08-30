package flow

import (
	"github.com/rosewoodmedia/rwgo/lib/rwstrings"
)

// LabelledError is a quality.Error implementor which prepends a string to an
// error. LabelError is a human-readable error aggregator.
type LabelledError struct {
	Err   error
	Label string
	Flow  string
	Trace []byte
}

func LabelError(err error, label string) error {
	if err == nil {
		return nil
	}
	return LabelledError{
		Err:   err,
		Label: label,
		Flow:  "",
		Trace: nil,
	}
}

func (erro LabelledError) Error() string {
	return erro.Label + ": " + erro.Err.Error()
}

func (erro LabelledError) GetText() string {
	if innerErro, ok := erro.Err.(Error); ok {
		return erro.Label + ": " + innerErro.GetText()
	}
	errstr := erro.Label + ": " + erro.Err.Error() + "\n"
	if erro.Trace != nil {
		errstr += rwstrings.Indent(string(erro.Trace)) + "\n"
	}
	return errstr
}

func (erro LabelledError) GetFlow() string {
	return erro.Flow
}

func (erro LabelledError) GetLabel() string {
	return erro.Label
}
