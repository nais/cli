package flag

import "github.com/nais/cli/v2/internal/root"

type Aiven struct{ *root.Flags }

type Create struct {
	*Aiven
	Expire uint   `name:"expire" short:"e" usage:"Number of |DAYS| until the generated credentials expire."`
	Secret string `name:"secret" short:"s" usage:"|NAME| of the Kubernetes secret to store the credentials in. Will be generated if not provided."`
}

type CreateKafka struct {
	*Create
	Test int    `name:"test" short:"t" usage:"Create a test Kafka topic with the given |NAME|."`
	Pool string `name:"pool" short:"p" usage:"The |NAME| of the pool to create the Kafka instance in."`
}

type CreateOpenSearch struct {
	*Create
	Instance string `name:"instance" short:"i" usage:"The name of the OpenSearch |INSTANCE|."`
	Access   string `name:"access" short:"a" usage:"The access |LEVEL|."`
}
