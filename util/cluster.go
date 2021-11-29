package util

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hazelcast/hazelcast-cloud-sdk-go/models"
	"github.com/jedib0t/go-pretty/v6/list"
)

func  AugmentStarterClusterType(starterClusterCreateClusterType string) (models.StarterClusterType, error) {
	switch strings.ToUpper(starterClusterCreateClusterType) {
	case "FREE":
		return models.Free, nil
	case "SMALL":
		return models.Small, nil
	case "MEDIUM":
		return models.Medium, nil
	case "LARGE":
		return models.Large, nil
	default:
		return "", errors.New("you can only select FREE, SMALL, MEDIUM or LARGE for cluster type")
	}
}

func AugmentZoneType(enterpriseClusterCreateZoneType string) (models.ZoneType, error) {
	switch enterpriseClusterCreateZoneType {
	case "SINGLE":
		return models.ZoneTypeSingle, nil
	case "MULTI":
		return models.ZoneTypeMultiple, nil
	default:
		return "", errors.New("you can only select SINGLE or MULTI as a zone type")
	}
}

func getEnabledDisable(bool bool) string {
	if bool {
		return "Enabled"
	}
	return "Disabled"
}

func getIsClusterEnterprise(cluster models.Cluster) bool {
	return cluster.ProductType.Name == models.Enterprise
}

func printCluster(cluster models.Cluster, printStyle PrintStyle) {
	wr := list.NewWriter()
	wr.SetOutputMirror(os.Stdout)
	wr.SetStyle(list.StyleConnectedBold)

	wr.AppendItem(fmt.Sprintf("Id: %s", cluster.Id))
	wr.AppendItem(fmt.Sprintf("Name: %s", cluster.Name))
	wr.AppendItem(fmt.Sprintf("Release Name: %s", cluster.ReleaseName))
	wr.AppendItem(fmt.Sprintf("Customer Id: %d", cluster.CustomerId))
	wr.AppendItem(fmt.Sprintf("Password: %s", cluster.Password))
	wr.AppendItem(fmt.Sprintf("Port: %d", cluster.Port))
	wr.AppendItem(fmt.Sprintf("Hazelcast Version: %s", cluster.HazelcastVersion))
	wr.AppendItem(fmt.Sprintf("Auto Scaling: %s", getEnabledDisable(cluster.IsAutoScalingEnabled)))
	wr.AppendItem(fmt.Sprintf("Hot Backup: %s", getEnabledDisable(cluster.IsHotBackupEnabled)))
	wr.AppendItem(fmt.Sprintf("Hot Restart: %s", getEnabledDisable(cluster.IsHotRestartEnabled)))
	wr.AppendItem(fmt.Sprintf("IP Whitelist: %s", getEnabledDisable(cluster.IsIpWhitelistEnabled)))
	wr.AppendItem(fmt.Sprintf("TLS: %s", getEnabledDisable(cluster.IsTlsEnabled)))

	wr.AppendItem("Product Type")
	wr.Indent()
	wr.AppendItem(fmt.Sprintf("Name: %s", cluster.ProductType.Name))
	wr.AppendItem(fmt.Sprintf("Free: %t", cluster.ProductType.IsFree))
	wr.UnIndent()

	wr.AppendItem(fmt.Sprintf("State: %s", cluster.State))
	wr.AppendItem(fmt.Sprintf("Created at: %s", cluster.CreatedAt))
	wr.AppendItem(fmt.Sprintf("Started at: %s", cluster.StartedAt))
	wr.AppendItem(fmt.Sprintf("Stopped at: %s", cluster.StoppedAt))

	wr.AppendItem("Cloud Provider")
	wr.Indent()
	wr.AppendItem(fmt.Sprintf("Name: %s", cluster.CloudProvider.Name))
	wr.AppendItem(fmt.Sprintf("Region: %s", cluster.CloudProvider.Region))
	if getIsClusterEnterprise(cluster) {
		wr.AppendItem(fmt.Sprintf("Availability Zones: %s", strings.Join(cluster.CloudProvider.AvailabilityZones, ", ")))
	}
	wr.UnIndent()

	wr.AppendItem("Discovery Tokens")
	wr.Indent()
	for _, discoveryToken := range cluster.DiscoveryTokens {
		wr.AppendItem(fmt.Sprintf("Source: %s", discoveryToken.Source))
		wr.AppendItem(fmt.Sprintf("Token: %s", discoveryToken.Token))
	}
	wr.UnIndent()

	wr.AppendItem("Specs")
	wr.Indent()
	wr.AppendItem(fmt.Sprintf("Total Memory(GiB): %.1f ", cluster.Specs.TotalMemory))
	if getIsClusterEnterprise(cluster) {
		wr.AppendItem(fmt.Sprintf("Instance Type: %s", cluster.Specs.InstanceType))
		wr.AppendItem(fmt.Sprintf("Instance Per Zone: %d", cluster.Specs.InstancePerZone))
		wr.AppendItem(fmt.Sprintf("Native Memory: %d", cluster.Specs.NativeMemory))
		wr.AppendItem(fmt.Sprintf("Heap Memory: %d", cluster.Specs.HeapMemory))
		wr.AppendItem(fmt.Sprintf("Cpu: %d", cluster.Specs.Cpu))
	}
	wr.UnIndent()

	if getIsClusterEnterprise(cluster) {
		wr.AppendItem("Networking")
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Type: %s ", cluster.Networking.Type))
		wr.AppendItem(fmt.Sprintf("Cidr Block: %s", cluster.Networking.CidrBlock))
		wr.AppendItem("Peering")
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Status: %s", getEnabledDisable(cluster.Networking.Peering.IsEnabled)))
		wr.UnIndent()
		wr.AppendItem("Private Link")
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Url: %s", cluster.Networking.PrivateLink.Url))
		wr.AppendItem(fmt.Sprintf("State: %s", cluster.Networking.PrivateLink.State))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("Data Structures")
	wr.Indent()
	wr.AppendItem("Map Configs")
	for _, mapConfig := range cluster.DataStructures.MapConfigs {
		wr.Indent()
		wr.AppendItem(mapConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", mapConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", mapConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Eviction Policy: %s", mapConfig.EvictionPolicy))
		if len(mapConfig.MapIndices) != 0 {
			wr.AppendItem("Map Indexes:")
			for _, mapIndex := range mapConfig.MapIndices {
				wr.Indent()
				wr.AppendItem(fmt.Sprintf("Name: %s", mapIndex.Name))
				wr.UnIndent()
			}
		}
		if mapConfig.MapStore.ClassName != "" {
			wr.AppendItem("Map Store")
			wr.Indent()
			wr.AppendItem(fmt.Sprintf("Class Name: %s", mapConfig.MapStore.ClassName))
			wr.AppendItem(fmt.Sprintf("Write Batch Size: %d", mapConfig.MapStore.WriteBatchSize))
			wr.AppendItem(fmt.Sprintf("Write Colaescing: %t", mapConfig.MapStore.WriteCoalescing))
			wr.AppendItem(fmt.Sprintf("Initial Load Mode: %s", mapConfig.MapStore.InitialLoadMode))
			wr.AppendItem(fmt.Sprintf("Write Delay Seconds: %d", mapConfig.MapStore.WriteDelaySeconds))
			wr.UnIndent()
		}
		wr.AppendItem(fmt.Sprintf("Max Idle Seconods: %d", mapConfig.MaxIdleSeconds))
		wr.AppendItem(fmt.Sprintf("Max Size: %d", mapConfig.MaxSize))
		wr.AppendItem(fmt.Sprintf("Max Size Policy: %s", mapConfig.MaxSizePolicy))
		wr.AppendItem(fmt.Sprintf("Ttl Seconds: %d", mapConfig.TtlSeconds))
		wr.AppendItem(fmt.Sprintf("Ready: %t", mapConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("List Configs:")
	for _, listConfig := range cluster.DataStructures.ListConfigs {
		wr.Indent()
		wr.AppendItem(listConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", listConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", listConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Max Size: %d", listConfig.MaxSize))
		wr.AppendItem(fmt.Sprintf("Ready: %t", listConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("Set Configs:")
	for _, setConfig := range cluster.DataStructures.SetConfigs {
		wr.Indent()
		wr.AppendItem(setConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", setConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", setConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Max Size: %d", setConfig.MaxSize))
		wr.AppendItem(fmt.Sprintf("Ready: %t", setConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("Queue Configs:")
	for _, queueConfig := range cluster.DataStructures.QueueConfigs {
		wr.Indent()
		wr.AppendItem(queueConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", queueConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", queueConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Empty Queue TTL: %d", queueConfig.EmptyQueueTtl))
		wr.AppendItem(fmt.Sprintf("Max Size: %d", queueConfig.MaxSize))
		wr.AppendItem(fmt.Sprintf("Ready: %t", queueConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("Jcache Configs:")
	for _, jcacheConfig := range cluster.DataStructures.JCacheConfigs {
		wr.Indent()
		wr.AppendItem(jcacheConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", jcacheConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", jcacheConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Eviction Policy: %s", jcacheConfig.EvictionPolicy))
		wr.AppendItem(fmt.Sprintf("Key Type: %s", jcacheConfig.KeyType))
		wr.AppendItem(fmt.Sprintf("Value Type: %s", jcacheConfig.ValueType))
		wr.AppendItem(fmt.Sprintf("Max Size: %d", jcacheConfig.MaxSize))
		wr.AppendItem(fmt.Sprintf("Max Size Policy: %s", jcacheConfig.MaxSizePolicy))
		wr.AppendItem(fmt.Sprintf("TTL Seconds: %d", jcacheConfig.TtlSeconds))
		wr.AppendItem(fmt.Sprintf("Ready: %t", jcacheConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("MultiMap Configs:")
	for _, multiMapConfig := range cluster.DataStructures.MultiMapConfigs {
		wr.Indent()
		wr.AppendItem(multiMapConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", multiMapConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", multiMapConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Value Collection Type: %s", multiMapConfig.ValueCollectionType))
		wr.AppendItem(fmt.Sprintf("Ready: %t", multiMapConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("Topic Configs:")
	for _, topicConfig := range cluster.DataStructures.TopicConfigs {
		wr.Indent()
		wr.AppendItem(topicConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Global Ordering: %s", getEnabledDisable(topicConfig.GlobalOrdering)))
		wr.AppendItem(fmt.Sprintf("Ready: %t", topicConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("RingBuffer Configs:")
	for _, ringBufferConfig := range cluster.DataStructures.RingBufferConfigs {
		wr.Indent()
		wr.AppendItem(ringBufferConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Async Backup Count: %d", ringBufferConfig.AsyncBackupCount))
		wr.AppendItem(fmt.Sprintf("Backup Count: %d", ringBufferConfig.BackupCount))
		wr.AppendItem(fmt.Sprintf("Capacity: %d", ringBufferConfig.Capacity))
		wr.AppendItem(fmt.Sprintf("In Memory Formay: %s", ringBufferConfig.InMemoryFormat))
		wr.AppendItem(fmt.Sprintf("TTL Seconds: %d", ringBufferConfig.TtlSeconds))
		wr.AppendItem(fmt.Sprintf("Ready: %t", ringBufferConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("ReliableTopic Configs:")
	for _, reliableTopicConfig := range cluster.DataStructures.ReliableTopicConfigs {
		wr.Indent()
		wr.AppendItem(reliableTopicConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("Topic Overload Policy: %s", reliableTopicConfig.TopicOverloadPolicy))
		wr.AppendItem(fmt.Sprintf("Read Batch Size: %d", reliableTopicConfig.ReadBatchSize))
		wr.AppendItem(fmt.Sprintf("Ready: %t", reliableTopicConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.AppendItem("ReliableTopic Configs:")
	for _, replicatedMapConfig := range cluster.DataStructures.ReplicatedMapConfigs {
		wr.Indent()
		wr.AppendItem(replicatedMapConfig.Name)
		wr.Indent()
		wr.AppendItem(fmt.Sprintf("In Memory Format: %s", replicatedMapConfig.InMemoryFormat))
		wr.AppendItem(fmt.Sprintf("Async Fill Up: %s", getEnabledDisable(replicatedMapConfig.AsyncFillUp)))
		wr.AppendItem(fmt.Sprintf("Ready: %t", replicatedMapConfig.IsReady))
		wr.UnIndent()
		wr.UnIndent()
	}

	wr.UnIndent()
	wr.UnIndent()
	if printStyle == PrintStyleDefault {
		wr.Render()
	} else if printStyle == PrintStyleHtml {
		wr.RenderHTML()
	} else if printStyle == PrintStyleMarkdown {
		wr.RenderMarkdown()
	} else if printStyle == PrintStyleCsv {
		color.Red("CSV not supported on this type.")
		os.Exit(1)
	}
}
