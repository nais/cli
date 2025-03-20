package kubeconfig

type filterOptions struct {
	fromScratch       bool
	includeCi         bool
	includeManagement bool
	includeOnprem     bool
	overwrite         bool
	verbose           bool
	prefixWithTenants bool
	excludeClusters   []string
}

type FilterOption func(options *filterOptions)

func WithFromScratch(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.fromScratch = enabled
	}
}

// WithCiClusters is used by Narc
func WithCiClusters(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeCi = include
	}
}

// WithManagementClusters is used by Narc
func WithManagementClusters(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeManagement = include
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

func WithOverwriteData(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.overwrite = enabled
	}
}

// WithPrefixedTenants is used by Narc
func WithPrefixedTenants(prefix bool) FilterOption {
	return func(options *filterOptions) {
		options.prefixWithTenants = prefix
	}
}

func WithVerboseLogging(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.verbose = enabled
	}
}

var DefaultFilterOptions = []FilterOption{}
