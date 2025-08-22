package flag

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
)

type OpenSearch struct {
	*alpha.Alpha
}

type Create struct {
	*OpenSearch
}

type Delete struct {
	*OpenSearch
}

type Describe struct {
	*OpenSearch
}

type List struct {
	*OpenSearch
}

type Update struct {
	*OpenSearch
}
