package command

import (
	"fmt"
	"io"
	"time"
)

// accessSpinner serializes terminal writes so completion never races with an animated frame.
type accessSpinner struct {
	done    chan struct{}
	stopped chan struct{}
}

func newAccessSpinner(w io.Writer) *accessSpinner {
	s := &accessSpinner{done: make(chan struct{}), stopped: make(chan struct{})}
	go s.run(w)
	return s
}

func (s *accessSpinner) run(w io.Writer) {
	defer close(s.stopped)
	const clearLine = "\r\x1b[2K"
	frames := [...]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	frame := 0
	draw := func() {
		_, _ = fmt.Fprintf(w, "%s%s Preparing personal Postgres access... (%s)", clearLine, frames[frame%len(frames)], time.Since(start).Round(time.Second))
		frame++
	}
	draw()
	for {
		select {
		case <-ticker.C:
			draw()
		case <-s.done:
			_, _ = io.WriteString(w, clearLine)
			return
		}
	}
}

func (s *accessSpinner) Stop() {
	close(s.done)
	<-s.stopped
}
