package cmd

import (
	"context"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
	"github.com/hazelcast/hazelcast-cloud-cli/util"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var availabilityZoneCloudProvider string
var availabilityZoneRegion string
var availabilityZoneInstanceType string
var availabilityZoneInstanceCount int

var availabilityZoneCmd = &cobra.Command{
	Use:     "availability-zone",
	Aliases: []string{"az"},
	Short:   "This command used to collect supported availability zone related information according to needs.",
}

var availabilityZoneListCmd = &cobra.Command{
	Use:     "list",
	Short:   "This command shows a list of availability zones.",
	Long:    "This command shows a list of availability zones. Availability zones depend on the cloud provider, region, instance type and count.",
	Example: "hzcloud availability-zone list --cloud-provider=aws --region=us-west-2 --instance-type=m5.large --count=3",
	Run: func(cmd *cobra.Command, args []string) {
		client := internal.NewClient()
		availabilityZones := internal.Validate(client.AvailabilityZone.List(context.Background(), &models.AvailabilityZoneInput{
			CloudProvider: availabilityZoneCloudProvider,
			Region:        availabilityZoneRegion,
			InstanceType:  availabilityZoneInstanceType,
			InstanceCount: availabilityZoneInstanceCount,
		})).(*[]models.AvailabilityZone)
		header := table.Row{"#", "Name"}
		rows := []table.Row{}
		for k, availabilityZone := range *availabilityZones {
			rows = append(rows, table.Row{k + 1, availabilityZone.Name})
		}
		util.Print(util.PrintRequest{
			Data:       availabilityZones,
			Header:     header,
			Rows:       rows,
			PrintStyle: util.PrintStyle(outputStyle),
		})
	},
	RunE:                       nil,
	PostRun:                    nil,
	PostRunE:                   nil,
	PersistentPostRun:          nil,
	PersistentPostRunE:         nil,
	SilenceErrors:              false,
	SilenceUsage:               false,
	DisableFlagParsing:         false,
	DisableAutoGenTag:          false,
	DisableFlagsInUseLine:      false,
	DisableSuggestions:         false,
	SuggestionsMinimumDistance: 0,
	TraverseChildren:           false,
	FParseErrWhitelist:         cobra.FParseErrWhitelist{},
}

func init() {
	rootCmd.AddCommand(availabilityZoneCmd)
	availabilityZoneCmd.AddCommand(availabilityZoneListCmd)
	availabilityZoneListCmd.Flags().StringVar(&availabilityZoneCloudProvider, "cloud-provider", "", "name of the cloud provider")
	availabilityZoneListCmd.Flags().StringVar(&availabilityZoneRegion, "region", "", "name of the cegion")
	availabilityZoneListCmd.Flags().StringVar(&availabilityZoneInstanceType, "instance-type", "", "name of the instance type")
	availabilityZoneListCmd.Flags().IntVar(&availabilityZoneInstanceCount, "count", 0, "instance count per zone")
	if err := availabilityZoneListCmd.MarkFlagRequired("cloud-provider"); err != nil {
		panic(err)
	}
	if err := availabilityZoneListCmd.MarkFlagRequired("region"); err != nil {
		panic(err)
	}
	if err := availabilityZoneListCmd.MarkFlagRequired("instance-type"); err != nil {
		panic(err)
	}
	if err := availabilityZoneListCmd.MarkFlagRequired("count"); err != nil {
		panic(err)
	}
}
