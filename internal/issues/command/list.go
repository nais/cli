package command

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/nais/cli/internal/issues"
	"github.com/nais/cli/internal/issues/command/flag"
	"github.com/nais/naistrix"
	"github.com/pterm/pterm"
	"golang.org/x/term"
)

type filters struct {
	issueType    string
	severity     string
	environment  string
	resourceName string
	resourceType string
}

func listIssues(parentFlags *flag.Issues) *naistrix.Command {
	flags := &flag.List{Issues: parentFlags}
	return &naistrix.Command{
		Name:        "list",
		Title:       "List issues.",
		Description: "This command lists all issues for a given team.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "team"},
		},
		ValidateFunc: func(ctx context.Context, args *naistrix.Arguments) error {
			if args.Get("team") == "" {
				return fmt.Errorf("team cannot be empty")
			}
			return nil
		},
		Examples: []naistrix.Example{
			{
				Description: "List all issues for the team named my-team.",
				Command:     "my-team",
			},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			issues, err := issues.GetAll(ctx, args.Get("team"))
			// if flags.Filter != "" {
			// 	filters, err := parseFilter(flags.Filter)
			// 	if err != nil {
			// 		return fmt.Errorf("parse filter: %w", err)
			// 	}
			// }

			if err != nil {
				return fmt.Errorf("fetching issues: %w", err)
			}

			data := pterm.TableData{
				{
					"Issue",
					"Severity",
					"Resource Name",
					"Resource Type",
					"Environment",
					"Message",
				},
			}

			width, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				fmt.Println("could not get terminal size:", err)
				width = 160
			}

			for _, i := range issues {
				data = append(data, []string{
					i.IssueType,
					i.Severity,
					i.ResourceName,
					i.ResourceType,
					i.Environment,
					truncateString(i.Message, width-100),
				})
			}
			return pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(data).Render()
		},
	}
}

func parseFilter(s string) (*flag.Filters, error) {

	ret := &flag.Filters{}
	parts := strings.Split(s, ",")
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("incorrect filter: %s", part)
		}
		key, value := kv[0], kv[1]
		switch key {
		case "environment":
			ret.Environment = value
		case "severity":
			ret.Severity = value
		case "resourcename":
			ret.ResourceName = value
		case "resourcetype":
			ret.ResourceType = value
		case "issuetype":
			ret.IssueType = value
		default:
			return nil, fmt.Errorf("unknown filter key: %s", key)
		}

	}
	return ret, nil
}

func truncateString(str string, max int) string {
	truncated := ""
	count := 0
	if len(str) < max {
		return str
	}

	for _, char := range str {
		truncated += string(char)
		count++
		if count >= max {
			break
		}
	}
	return truncated + "[...]"
}
