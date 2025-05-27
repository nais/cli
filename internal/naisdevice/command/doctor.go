package command

import (
	"context"

	"github.com/mitchellh/go-ps"
	"github.com/nais/cli/internal/cli"
	doc "github.com/nais/cli/internal/doctor"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func doctorcmd(_ *root.Flags) *cli.Command {
	return cli.NewCommand("doctor", "Check the health of your naisdevice.",
		cli.WithRun(func(_ context.Context, w output.Output, _ []string) error {
			results := examination(w).Run()
			for key, value := range results {
				w.Printf("%s ", key)
				if value.Result == doc.OK {
					w.Println("✅")
				} else {
					w.Printf("❌ (%s)\n", value.ErrMsg)
				}
			}
			w.Println()
			return nil
		}),
	)
}

func examination(w output.Output) doc.Examination {
	checkName := "Is Kolide and Osquery running?"
	return doc.Examination{
		Name: "Device checks",
		Checks: []doc.Check{
			{
				Name:   checkName,
				Worker: kolideWorker(checkName, w),
			},
		},
	}
}

func kolideWorker(checkName string, w output.Output) doc.Worker {
	return func() doc.CheckReport {
		kolideRunning, err := isRunning("launcher", w)
		if err != nil || !kolideRunning {
			return errorReport(checkName, "Kolide is not running")
		}
		osQueryRunning, err := isRunning("osqueryd", w)
		if err != nil || !osQueryRunning {
			return errorReport(checkName, "Osquery is not running")
		}
		return okReport(checkName)
	}
}

func isRunning(desiredProc string, w output.Output) (bool, error) {
	runningProcs, err := ps.Processes()
	if err != nil {
		w.Printf("Process listing failed: %v\n", err)
		return false, err
	}
	for _, runningProc := range runningProcs {
		if desiredProc == runningProc.Executable() {
			return true, nil
		}
	}
	return false, nil
}

func errorReport(checkName, errMsg string) doc.CheckReport {
	return doc.CheckReport{
		CheckName: checkName,
		Result:    doc.Error,
		ErrMsg:    errMsg,
	}
}

func okReport(checkName string) doc.CheckReport {
	return doc.CheckReport{
		CheckName: checkName,
		Result:    doc.OK,
		ErrMsg:    "",
	}
}
