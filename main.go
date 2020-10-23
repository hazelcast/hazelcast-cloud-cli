package main

import (
	"github.com/hazelcast/hazelcast-cloud-cli/cmd"
	"github.com/hazelcast/hazelcast-cloud-cli/internal"
)

func main() {
	cmd.Execute()
	updaterService := internal.NewUpdaterService()
	updaterService.Clean()
	updaterService.Check(false)
}
