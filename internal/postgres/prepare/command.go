package prepare

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
)

func Run(ctx context.Context, applicationName string, flags *flag.Prepare) error {
	fmt.Print("\nAre you sure you want to continue (y/N): ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	if !strings.EqualFold(strings.TrimSpace(input.Text()), "y") {
		return fmt.Errorf("cancelled by user")
	}

	return postgres.PrepareAccess(ctx, applicationName, flags.Namespace, flags.Context, flags.Schema, flags.AllPrivileges)
}
