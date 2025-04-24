package doctor

import (
	"testing"
	"time"
)

func TestAllChecksAreRun(t *testing.T) {
	examination := okExaminationWith2SecondWorkers()
	res := examination.Run()
	if len(res) != len(examination.Checks) {
		t.Fatalf("nr of results should not differ from the nr of checks")
	}
}

func TestChecksAreRunConcurrently(t *testing.T) {
	start := time.Now()
	_ = okExaminationWith2SecondWorkers().Run()
	elapsed := time.Since(start)
	if elapsed >= time.Second*3 {
		t.Fatalf("checks took to long, they are probably not run concurrently")
	}
}

func okExaminationWith2SecondWorkers() Examination {
	return Examination{
		Name: "Bogus Examination",
		Checks: []Check{
			{
				Name:   "First check",
				Worker: twoSecondWorker("Bogus Examination worker 1"),
			},
			{
				Name:   "Second check",
				Worker: twoSecondWorker("Bogus Examination worker 2"),
			},
		},
	}
}

func twoSecondWorker(checkName string) Worker {
	return func() CheckReport {
		time.Sleep(time.Second * 2)
		return CheckReport{
			CheckName: checkName,
			Result:    OK,
			ErrMsg:    "",
		}
	}
}
