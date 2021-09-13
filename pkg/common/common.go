package common

import (
	"fmt"
	"io/ioutil"
)

const (
	FilePermission = 0775
)

func Destination(dest, filename string) string {
	return fmt.Sprintf("%s/%s", dest, filename)
}

func WriteToFile(dest, filename string, value []byte) error {
	err := ioutil.WriteFile(Destination(dest, filename), value, FilePermission)
	if err != nil {
		return err
	}
	return nil
}
