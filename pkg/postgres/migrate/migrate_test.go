package migrate_test

import (
	"github.com/nais/cli/pkg/option"
	"github.com/nais/cli/pkg/postgres/migrate"
	"github.com/nais/cli/pkg/postgres/migrate/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command", func() {
	var cfg config.Config
	const cmd = migrate.CommandSetup

	Context("JobName", func() {
		DescribeTableSubtree("given configuration", func(mutateFn func(), expected string) {
			BeforeEach(func() {
				cfg = config.Config{
					AppName:   "some-app",
					Namespace: "test-namespace",
					Target: config.InstanceConfig{
						InstanceName: option.Some("target-instance"),
					},
				}
				mutateFn()
			})

			It("should generate a job name suffixed with the command", func() {
				actual := cmd.JobName(cfg)
				Expect(len(actual)).To(BeNumerically("<=", 63))
				Expect(actual).To(Equal(expected))
			})
		},
			Entry("happy path with reasonable lengths for app and instance",
				func() {},
				"migration-some-app-target-instance-setup",
			),
			Entry("very long app name",
				func() {
					cfg.AppName = "some-unnecessarily-long-app-name-that-should-be-truncated"
				},
				"migration-some-unnecessarily-long-app-name-that--eb4938d8-setup",
			),
			Entry("very long instance name",
				func() {
					cfg.Target.InstanceName = option.Some("some-unnecessarily-long-instance-name-that-should-be-truncated")
				},
				"migration-some-app-some-unnecessarily-long-insta-63093bcb-setup",
			),
		)
	})
})
