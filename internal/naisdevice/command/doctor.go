package command

import (
	"context"

	"github.com/mitchellh/go-ps"
	doc "github.com/nais/cli/v2/internal/doctor"
	"github.com/nais/cli/v2/internal/root"
	"github.com/nais/naistrix"
)

func doctorcmd(_ *root.Flags) *naistrix.Command {
	return &naistrix.Command{
		Name:  "doctor",
		Title: "Check the health of your naisdevice.",
		RunFunc: func(_ context.Context, out naistrix.Output, _ []string) error {
			results := examination(out).Run(out)
			for key, value := range results {
				out.Printf("%s ", key)
				if value.Result == doc.OK {
					out.Println("✅")
				} else {
					out.Printf("❌ (%s)\n", value.ErrMsg)
				}
			}
			out.Println()
			return nil
		},
	}
}

func examination(out naistrix.Output) doc.Examination {
	checkName := "Is Kolide and Osquery running?"
	return doc.Examination{
		Name: "Device checks",
		Checks: []doc.Check{
			{
				Name:   checkName,
				Worker: kolideWorker(checkName, out),
			},
		},
	}
}

func kolideWorker(checkName string, out naistrix.Output) doc.Worker {
	return func() doc.CheckReport {
		kolideRunning, err := isRunning("launcher", out)
		if err != nil || !kolideRunning {
			return errorReport(checkName, "Kolide is not running")
		}
		osQueryRunning, err := isRunning("osqueryd", out)
		if err != nil || !osQueryRunning {
			return errorReport(checkName, "Osquery is not running")
		}
		return okReport(checkName)
	}
}

func isRunning(desiredProc string, out naistrix.Output) (bool, error) {
	runningProcs, err := ps.Processes()
	if err != nil {
		out.Printf("Process listing failed: %v\n", err)
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
