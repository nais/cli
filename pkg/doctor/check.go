package doctor

import "context"

var checks []Check

func AddCheck(check Check) {
	checks = append(checks, check)
}

type Check interface {
	Name() string
	Help() string
	Check(ctx context.Context, cfg *Config) []error
}

// Ackable is implemented by checks that has to be acknowledged by the user.
type Ackable interface {
	Ack()
}
