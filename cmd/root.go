package cmd

import (
	"fmt"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/spf13/cobra"
	"os"
)

var outputStyle string

var rootCmd = &cobra.Command{
	Use:   "hzcloud",
	Short: "hzcloud is a command line interface (CLI) for the Hazelcast Cloud API.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputStyle, "output", "o", "default", "--output json")
	internal.NewConfigService().CreateConfig()
}
