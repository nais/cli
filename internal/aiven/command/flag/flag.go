package flag

import "github.com/nais/cli/internal/root"

type Aiven struct{ *root.Flags }

type Create struct {
	*Aiven
	Expire uint
	Secret string
}

type CreateKafka struct {
	*Create
	Test int
	Pool string
}

type CreateOpenSearch struct {
	*Create
	Instance string
	Access   string
}
