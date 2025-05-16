package root

type Flags struct {
	VerboseLevel int
}

func (f *Flags) IsVerbose() bool {
	return f != nil && f.VerboseLevel > 0
}

func (f *Flags) IsDebug() bool {
	return f != nil && f.VerboseLevel > 1
}
