package main

import (
	"github.com/hazelcast/hazelcast-cloud-cli/cmd"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
)

func main() {
	updaterService := internal.NewUpdaterService()
	updaterService.Clean()
	cmd.Execute()
	updaterService.Check(false)
}
