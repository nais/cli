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

	ListenAddr string `name:"listen" short:"l" usage:"Address the proxy will listen on."`
}

type Teams struct {
	*Api

	All    bool   `name:"all" short:"a" usage:"List all teams, not just the ones you are a member of"`
	Output string `name:"output" short:"o" usage:"Format output (table|json)"`
}

type Schema struct {
	*Api
}
