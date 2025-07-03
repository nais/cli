package application

import (
	"fmt"
	"net/url"
	"os"
	"runtime/debug"
	"strings"

	"github.com/nais/cli/internal/urlopen"
	"github.com/nais/cli/internal/version"
)

func fmtCode(c string) string {
	return fmt.Sprintf("`%s`", c)
}

func fmtMultilineCode(c string) string {
	return fmt.Sprintf("```\n%s\n```", c)
}

func handlePanic(recoveredFrom any) {
	recoveredString := fmt.Sprintf("%v", recoveredFrom)

	fmt.Printf("Unexpected error occurred: %v\nstack:\n %s", recoveredString, debug.Stack())
	fmt.Println("")
	fmt.Println("We would appreciate if you create an issue on GitHub.")
	fmt.Print("Would you like to open a browser with a pre-filled issue? (check for sensitive information) [y/N] ")

	var response string
	_, _ = fmt.Scanln(&response)
	if strings.EqualFold(response, "y") {
		body := fmt.Sprintf(`Command ran: %s

Error: %s
Stack trace:
%s`,
			fmtCode(strings.Join(os.Args, " ")),
			fmtCode(recoveredString),
			fmtMultilineCode(string(debug.Stack())))

		u, _ := url.Parse("https://github.com/nais/cli/issues/new")
		values := u.Query()
		values.Add("title", fmt.Sprintf("Unexpected error in version %v", version.Version))
		values.Add("body", body)
		u.RawQuery = values.Encode()

		if err := urlopen.Open(u.String()); err != nil {
			fmt.Printf("Unable to open your browser, please open this manually: %s\n", u.String())
		}
	} else {
		fmt.Println("Skipping issue creation.")
	}
}
