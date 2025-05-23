package doctor

import (
	"context"
	"fmt"

	"github.com/mitchellh/go-ps"
	"github.com/nais/cli/internal/cli"
	doc "github.com/nais/cli/internal/doctor"
	"github.com/nais/cli/internal/root"
)

func Doctor(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("doctor", "Check the health of your naisdevice.",
		cli.WithHandler(run),
	)
}

func run(ctx context.Context, _ []string) error {
	results := examination().Run()
	for key, value := range results {
		fmt.Printf("%s ", key)
		if value.Result == doc.OK {
			println("✅")
		} else {
			fmt.Printf("❌ (%s)\n", value.ErrMsg)
		}
	}
	println()
	return nil
}

func examination() doc.Examination {
	checkName := "Is Kolide and Osquery running?"
	return doc.Examination{
		Name: "Device checks",
		Checks: []doc.Check{
			{
				Name:   checkName,
				Worker: kolideWorker(checkName),
			},
		},
	}
}

func kolideWorker(checkName string) doc.Worker {
	return func() doc.CheckReport {
		kolideRunning, err := isRunning("launcher")
		if err != nil || !kolideRunning {
			return errorReport(checkName, "Kolide is not running")
		}
		osQueryRunning, err := isRunning("osqueryd")
		if err != nil || !osQueryRunning {
			return errorReport(checkName, "Osquery is not running")
		}
		return okReport(checkName)
	}
}

func isRunning(desiredProc string) (bool, error) {
	runningProcs, err := ps.Processes()
	if err != nil {
		fmt.Printf("Process listing failed: %v\n", err)
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
