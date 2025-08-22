package task

// template
// _, _ = runTimedTask(d.ctx, d.flags.Timeout, "desc", func(ctx context.Context) (*any, error) {
// 	return nil, nil
// })

import (
	"context"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

func Timed[T any](parentCtx context.Context, timeout time.Duration, description string, f func(ctx context.Context) (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	done := make(chan string)
	lock := &sync.Mutex{}

	lock.Lock()
	go start(ctx, description, done, lock)
	ret, err := f(ctx)
	result := pterm.Green("done")
	if err != nil {
		result = pterm.Red("err")
	}
	stop(done, result, lock)
	return ret, err
}

func stop(done chan string, status string, lock *sync.Mutex) {
	if done != nil {
		select {
		case done <- status:
		default:
		}
		lock.Lock() // Ensure we wait for the status area to finish before returning
	} else {
		close(done)
		done = nil
	}
}

func start(ctx context.Context, description string, done chan string, lock *sync.Mutex) {
	defer lock.Unlock()
	statusArea, err := (&pterm.AreaPrinter{}).Start()
	if err != nil {
		pterm.Error.Printf("failed to create status area: %v\n", err)
		return
	}
	defer statusArea.Stop()

	setStatus := func(status string) {
		statusArea.Update(pterm.Sprintf("%s [%v]", description, status))
	}

	for {
		select {
		case status := <-done:
			setStatus(status)
			pterm.Println()
			return
		case <-ctx.Done():
			switch err := ctx.Err(); err {
			case context.DeadlineExceeded:
				setStatus(pterm.Yellow("timeout"))
			case context.Canceled:
				setStatus(pterm.Red("cancelled"))
			}
			pterm.Println()
			return
		case <-time.After(50 * time.Millisecond):
			if deadline, ok := ctx.Deadline(); ok {
				setStatus(time.Until(deadline).Round(time.Second).String())
			} else {
				setStatus("?")
			}
		}
	}
}
