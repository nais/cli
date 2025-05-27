package root

type Flags struct {
	// VerboseLevel indicates the verbosity level of Nais CLI.
	VerboseLevel int
}

// IsVerbose checks if Nais CLI is running in verbose mode (-v).
func (f *Flags) IsVerbose() bool {
	return f != nil && f.VerboseLevel > 0
}

// IsDebug checks if Nais CLI is running in debug mode (-vv or higher).
func (f *Flags) IsDebug() bool {
	return f != nil && f.VerboseLevel > 1
}
