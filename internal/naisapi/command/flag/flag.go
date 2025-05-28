package flag

import "github.com/nais/cli/internal/root"

type Alpha struct {
	*root.Flags
}

type Api struct {
	*Alpha
}

type Proxy struct {
	*Api

	ListenAddr string
}

type Teams struct {
	*Api
}

type Schema struct {
	*Api
}
