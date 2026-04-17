package flag

import (
	"github.com/nais/cli/internal/flags"
)

type Api struct {
	*flags.GlobalFlags
}

type Proxy struct {
	*Api
	ListenAddr string `name:"listen" short:"l" usage:"Address the proxy will listen on."`
}

type Schema struct {
	*Api
}
