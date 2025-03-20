package kubeconfig

type filterOptions struct {
	fromScratch       bool
	includeCi         bool
	includeKnada      bool
	includeManagement bool
	includeOnprem     bool
	overwrite         bool
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
		options.includeCi = include
	}
}

func WithKnadaCluster(include bool) FilterOption {
	return func(options *filterOptions) {
		options.includeKnada = include
	}
}

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

func WithVerboseLogging(enabled bool) FilterOption {
	return func(options *filterOptions) {
		options.verbose = enabled
	}
}

var DefaultFilterOptions = []FilterOption{}
