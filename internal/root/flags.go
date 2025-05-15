package root

type Flags struct {
	VerboseLevel int
}

func (f Flags) IsVerbose() bool {
	return f.VerboseLevel > 0
}

func (f Flags) IsDebug() bool {
	return f.VerboseLevel > 1
}
