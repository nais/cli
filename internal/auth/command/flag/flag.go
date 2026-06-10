package flag

import (
	"github.com/nais/cli/internal/flags"
)

type Auth struct {
	*flags.GlobalFlags
}

type Login struct {
	*Auth
	Nais bool `name:"nais" short:"n" usage:"Login using login.nais.io instead of gcloud."`
	Yes  bool `name:"yes" short:"y" usage:"Automatically answer yes to all prompts."`
}

type Logout struct {
	*Auth
	Nais bool `name:"nais" short:"n" usage:"Logout using login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\"."`
	Yes  bool `name:"yes" short:"y" usage:"Automatically answer yes to all prompts."`
}

type PrintAccessToken struct {
	*Auth
	Nais bool `name:"nais" short:"n" usage:"Print token from login.nais.io instead of gcloud.\nShould be used if you logged in using \"nais login --nais\"."`
}
