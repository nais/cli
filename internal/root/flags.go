package root

import "github.com/nais/cli/pkg/cli"

type Flags struct {
	// VerboseLevel indicates the verbosity level of Nais CLI.
	VerboseLevel cli.Count `name:"verbose" short:"v" usage:"Set verbosity level. Use -v for verbose, -vv for debug."`
}

// IsVerbose checks if Nais CLI is running in verbose mode (-v).
func (f *Flags) IsVerbose() bool {
	return f != nil && f.VerboseLevel > 0
}

// IsDebug checks if Nais CLI is running in debug mode (-vv).
func (f *Flags) IsDebug() bool {
	return f != nil && f.VerboseLevel > 1
}

// IsTrace checks if Nais CLI is running in trace mode (-vvv or higher).
func (f *Flags) IsTrace() bool {
	return f != nil && f.VerboseLevel > 2
}
