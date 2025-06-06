package command

import (
	"context"

	"github.com/mitchellh/go-ps"
	"github.com/nais/cli/internal/cli"
	doc "github.com/nais/cli/internal/doctor"
	"github.com/nais/cli/internal/root"
)

func doctorcmd(_ *root.Flags) *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Short: "Check the health of your naisdevice.",
		RunFunc: func(_ context.Context, out cli.Output, _ []string) error {
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

func examination(out cli.Output) doc.Examination {
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

func kolideWorker(checkName string, out cli.Output) doc.Worker {
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

func isRunning(desiredProc string, out cli.Output) (bool, error) {
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
