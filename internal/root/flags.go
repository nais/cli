package root

type Flags struct {
	VerboseLevel int
}

func (f Flags) IsVerbose() bool {
	return f.VerboseLevel > 0
}
