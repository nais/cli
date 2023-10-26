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
	excludeClusters   []string
}

type FilterOption func(options *filterOptions)

func WithFromScratch(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.fromScratch = enabled
	}
}

func WithCiClusters(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeOnprem = include
	}
}

func WithKnadaCluster(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeOnprem = include
	}
}

func WithManagementClusters(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeOnprem = include
	}
}

func WithOnpremClusters(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeOnprem = include
	}
}

func WithExcludeClusters(exclude []string) FilterOption {
	return func(options *filterOptions) {
		options.excludeClusters = exclude
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

func WithOverwriteData(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.overwrite = enabled
	}
}

func WithVerboseLogging(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.verbose = enabled
	}
}

var DefaultFilterOptions = []FilterOption{
	WithNAVPrefixSkipped(true),
}
