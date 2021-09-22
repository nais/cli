package config

type Config interface {
	WriteConfigToFile() error
	Set(key string, value []byte)
	Generate() (string, error)
}