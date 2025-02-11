package devicecmd

import (
	"fmt"

	"github.com/mitchellh/go-ps"
	"github.com/nais/cli/pkg/doctor"
	"github.com/urfave/cli/v2"
)

func doctorCommand() *cli.Command {
	return &cli.Command{
		Name:  "doctor",
		Usage: "Examine the health of your naisdevice",
		Action: func(context *cli.Context) error {
			results := examination().Run()
			for key, value := range results {
				fmt.Printf("%s ", key)
				if value.Result == doctor.OK {
					println("✅")
				} else {
					fmt.Printf("❌ (%s)\n", value.ErrMsg)
				}
			}
			println()
			return nil
		},
	}
}

func examination() doctor.Examination {
	checkName := "Is Kolide and Osquery running?"
	return doctor.Examination{
		Name: "Device checks",
		Checks: []doctor.Check{
			{
				Name:   checkName,
				Worker: kolideWorker(checkName),
			},
		},
	}
}

func kolideWorker(checkName string) doctor.Worker {
	return func() doctor.CheckReport {
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

func errorReport(checkName, errMsg string) doctor.CheckReport {
	return doctor.CheckReport{
		CheckName: checkName,
		Result:    doctor.Error,
		ErrMsg:    errMsg,
	}
}

func okReport(checkName string) doctor.CheckReport {
	return doctor.CheckReport{
		CheckName: checkName,
		Result:    doctor.OK,
		ErrMsg:    "",
	}
}
