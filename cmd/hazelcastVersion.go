package cmd

import (
	"context"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"strings"
)

var hazelcastVersionCmd = &cobra.Command{
	Use:     "hazelcast-version",
	Aliases: []string{"hv"},
	Short:   "This command used to collect supported available Hazelcast Versions on the cloud.",
}

var hazelcastVersionListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command lists available Hazelcast Versions.",
	Long:    "This command lists available Hazelcast Versions that Hazelcast Cloud support on their Enterprise product.",
	Example: "hzcloud hazelcast-version list",
	Args:    cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		versions := internal.Validate(client.HazelcastVersion.List(context.Background())).(*[]models.EnterpriseHazelcastVersion)
		header := table.Row{"#", "Version", "Upgradeable Versions"}
		rows := []table.Row{}
		for k, version := range *versions {
			rows = append(rows, table.Row{k + 1, version.Version, strings.Join(version.UpgradeableVersions, " ")})
		}
		util.Print(util.PrintRequest{
			Data:       versions,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
}

func init() {
	rootCmd.AddCommand(hazelcastVersionCmd)
	hazelcastVersionCmd.AddCommand(hazelcastVersionListCmd)
}
