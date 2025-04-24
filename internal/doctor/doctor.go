package doctor

import "fmt"

type Result int

const (
	OK    Result = 0
	Error Result = 1
)

type Worker func() CheckReport

type Check struct {
	Name   string
	Worker Worker
}

type CheckReport struct {
	CheckName string
	Result    Result
	ErrMsg    string
}

type Examination struct {
	Name   string
	Checks []Check
}

func (e Examination) Run() map[string]CheckReport {
	fmt.Printf("Running %d check(s)\n", len(e.Checks))
	resultQueue := make(chan CheckReport, len(e.Checks))
	results := make(map[string]CheckReport)
	for _, check := range e.Checks {
		c := check
		go func() {
			resultQueue <- c.Worker()
		}()
	}
	for range e.Checks {
		res := <-resultQueue
		results[res.CheckName] = res
	}
	close(resultQueue)
	return results
}
