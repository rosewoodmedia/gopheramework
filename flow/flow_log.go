package flow

import (
	"fmt"
	"time"
)

func (fl *Flow) LogWithTime(note string) {
	fl.Log(fmt.Sprintf("%s @%s", note, time.Now().UTC().Format(
		time.RFC3339,
	)))
}
