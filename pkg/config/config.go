package config

type Config interface {
	Write() error
	Finit() error
	Init()
	Set(key string, value []byte, destination string)
	Generate() error
}

const (
	ENV  = ".env"
	KCAT = "kcat"
	ALL  = "all"
)
