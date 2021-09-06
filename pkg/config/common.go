package config

import "fmt"

func Destination(dest, filename string) string {
	return fmt.Sprintf("%s/%s", dest, filename)
}
