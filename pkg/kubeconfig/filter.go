package kubeconfig

type filterOptions struct {
	fromScratch       bool
	includeCi         bool
	includeKnada      bool
	includeManagement bool
	includeOnprem     bool
	prefixWithTenant  bool
	overwrite         bool
	skipNAVPrefix     bool
	verbose           bool
}

type FilterOption func(options *filterOptions)

func WithOverwriteData(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.overwrite = enabled
	}
}

func WithFromScratch(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.fromScratch = enabled
	}
}

func WithIncludeOnprem(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeOnprem = include
	}
}

func WithVerboseLogging(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.verbose = enabled
	}
}

func WithNAVPrefixSkipped(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.skipNAVPrefix = enabled
	}
}

func WithPrefixedTenant(enbaled bool) FilterOption {
	return func(options *filterOptions) {
		options.prefixWithTenant = enbaled
	}
}

var DefaultFilterOptions = []FilterOption{
	WithNAVPrefixSkipped(true),
}
