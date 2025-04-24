package urlopen

import (
	"os/exec"
)

func Open(url string) error {
	return exec.Command("xdg-open", url).Run()
}
