package iostreams

import (
	"fmt"
	"io"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type spinnerState struct {
	stop chan struct{}
	done chan struct{}
}

// StartSpinner writes a progress indicator to ErrOut.
//
// When SpinnerEnabled is true the indicator is animated (one frame per 100ms).
// In CI mode, when CIRCLECI_SPINNER_DISABLED is set, or when the spinner is
// disabled for any other reason, msg is written as a plain-text line so
// progress is still visible in captured logs.
//
// The returned stop function must be called to clear the spinner and release
// the goroutine. It is safe to call more than once.
func (s *IOStreams) StartSpinner(msg string) func() {
	if !s.SpinnerEnabled {
		fmt.Fprintf(s.ErrOut, "%s\n", msg)
		return func() {}
	}

	state := &spinnerState{
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go runSpinner(s.ErrOut, msg, state)
	var stopped bool
	return func() {
		if !stopped {
			stopped = true
			close(state.stop)
			<-state.done
		}
	}
}

func runSpinner(w io.Writer, msg string, state *spinnerState) {
	defer close(state.done)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-state.stop:
			// Erase the spinner line.
			fmt.Fprint(w, "\r\033[K")
			return
		case <-ticker.C:
			fmt.Fprintf(w, "\r%s %s", spinnerFrames[i], msg)
			i = (i + 1) % len(spinnerFrames)
		}
	}
}
