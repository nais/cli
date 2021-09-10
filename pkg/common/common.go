package common

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/nais/debuk/pkg/application"
	"io/ioutil"
)

func Destination(dest, filename string) string {
	return fmt.Sprintf("%s/%s", dest, filename)
}

func WriteToFile(dest, filename, value string) error {
	if res, err := b64.StdEncoding.DecodeString(value); err == nil {
		err = ioutil.WriteFile(Destination(dest, filename), res, application.FilePermission)
		if err != nil {
			return err
		}
	}
	return nil
}
