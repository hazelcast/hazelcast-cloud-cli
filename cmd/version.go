package cmd

import (
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Aliases: []string{"v"},
	Short: "This command used to show version of Hazelcast Cloud CLI and update",
	Run: func(cmd *cobra.Command, args []string) {
		internal.NewUpdaterService().Check(true)
	},
}

var versionUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "This command updates Hazelcast Version.",
	Example: "hzcloud version update",
	Run: func(cmd *cobra.Command, args []string) {
		internal.NewUpdaterService().Update()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(versionUpdateCmd)
}

