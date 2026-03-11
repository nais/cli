package flag

import (
	"fmt"

	"github.com/nais/cli/internal/flags"
)

type Aiven struct {
	*flags.GlobalFlags
	Environment Environment `name:"environment" short:"e" usage:"The |ENVIRONMENT| to use."`
}

type Create struct {
	*Aiven
	Expire uint   `name:"expire" usage:"Number of |DAYS| until the generated credentials expire."`
	Secret string `name:"secret" short:"s" usage:"|NAME| of the Kubernetes secret to store the credentials in. Will be generated if not provided."`
}

type CreateKafka struct {
	*Create
	Test int    `name:"test" usage:"Create a test Kafka topic with the given |NAME|."`
	Pool string `name:"pool" short:"p" usage:"The |NAME| of the pool to create the Kafka instance in."`
}

type CreateOpenSearch struct {
	*Create
	Instance string `name:"instance" short:"i" usage:"The name of the OpenSearch |INSTANCE|."`
	Access   string `name:"access" short:"a" usage:"The access |LEVEL|."`
}

type GrantAccess struct {
	*Aiven
	Namespace string `name:"namespace" short:"n" usage:"REMOVED, see --team."`
}

func (g GrantAccess) UsesRemovedFlags() error {
	if g.Namespace != "" {
		return fmt.Errorf("the --namespace (-n) flag is replaced with the --team (-t) flag")
	}
	return nil
}

type GrantAccessStream struct {
	*GrantAccess
}

type GrantAccessTopic struct {
	*GrantAccess
	Access string `name:"access" short:"a" usage:"Access |LEVEL| (readwrite, read and write)."`
}
