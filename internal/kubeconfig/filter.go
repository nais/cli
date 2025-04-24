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

func WithVerboseLogging(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.verbose = enabled
	}
}

var DefaultFilterOptions = []FilterOption{}
