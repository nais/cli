package revoke

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/internal/postgres"
)

type Flags struct {
	*postgres.Flags
	Schema string
}

func Run(ctx context.Context, applicationName string, flags *Flags) error {
	fmt.Print("\nAre you sure you want to continue (y/N): ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
		return fmt.Errorf("cancelled by user")
	}

	return postgres.RevokeAccess(ctx, applicationName, flags.Namespace, flags.Context, flags.Schema)
}
