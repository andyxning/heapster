// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"fmt"
	"time"

	cadvisor "github.com/google/cadvisor/info/v1"
	"k8s.io/heapster/metrics/util/metrics"
)

const (
	CustomMetricPrefix = "custom/"
)

// Provided by Kubelet/cadvisor.
var StandardMetrics = []Metric{
	MetricUptime,
	MetricCpuUsage,
	MetricMemoryCache,
	MetricMemoryRSS,
	MetricMemoryUsage,
	MetricMemoryRSS,
	MetricMemoryCache,
	MetricMemoryWorkingSet,
	MetricMemoryPageFaults,
	MetricMemoryMajorPageFaults,
	MetricNetworkRx,
	MetricNetworkRxErrors,
	MetricNetworkTx,
	MetricNetworkTxErrors}

// Metrics computed based on cluster state using Kubernetes API.
var AdditionalMetrics = []Metric{
	MetricCpuRequest,
	MetricCpuLimit,
	MetricMemoryRequest,
	MetricMemoryLimit}

// Computed based on corresponding StandardMetrics.
var RateMetrics = []Metric{
	MetricCpuUsageRate,
	MetricNetworkRxRate,
	MetricNetworkTxRate,
	MetricDiskIOReadRate,
	MetricDiskIOWriteRate}

var RateMetricsMapping = map[string]Metric{
	MetricCpuUsage.MetricDescriptor.Name:    MetricCpuUsageRate,
	MetricNetworkRx.MetricDescriptor.Name:   MetricNetworkRxRate,
	MetricNetworkTx.MetricDescriptor.Name:   MetricNetworkTxRate,
	MetricDiskIORead.MetricDescriptor.Name:  MetricDiskIOReadRate,
	MetricDiskIOWrite.MetricDescriptor.Name: MetricDiskIOWriteRate}

var LabeledMetrics = []Metric{
	MetricDiskIORead,
	MetricDiskIOReadRate,
	MetricDiskIOWrite,
	MetricDiskIOWriteRate,
	MetricFilesystemUsage,
	MetricFilesystemLimit,
	MetricFilesystemAvailable,
	MetricFilesystemInodes,
	MetricFilesystemInodesFree,
	MetricAcceleratorMemoryTotal,
	MetricAcceleratorMemoryUsed,
	MetricAcceleratorDutyCycle,
}

var NodeAutoscalingMetrics = []Metric{
	MetricNodeCpuCapacity,
	MetricNodeMemoryCapacity,
	MetricNodeCpuAllocatable,
	MetricNodeMemoryAllocatable,
	MetricNodeCpuUtilization,
	MetricNodeMemoryUtilization,
	MetricNodeCpuReservation,
	MetricNodeMemoryReservation,
}

var CpuMetrics = []Metric{
	MetricCpuLimit,
	MetricCpuRequest,
	MetricCpuUsage,
	MetricCpuUsageRate,
	MetricNodeCpuAllocatable,
	MetricNodeCpuCapacity,
	MetricNodeCpuReservation,
	MetricNodeCpuUtilization,
}
var FilesystemMetrics = []Metric{
	MetricFilesystemAvailable,
	MetricFilesystemLimit,
	MetricFilesystemUsage,
	MetricFilesystemInodes,
	MetricFilesystemInodesFree,
}
var MemoryMetrics = []Metric{
	MetricMemoryLimit,
	MetricMemoryMajorPageFaults,
	MetricMemoryPageFaults,
	MetricMemoryRequest,
	MetricMemoryUsage,
	MetricMemoryRSS,
	MetricMemoryCache,
	MetricMemoryWorkingSet,
	MetricNodeMemoryAllocatable,
	MetricNodeMemoryCapacity,
	MetricNodeMemoryUtilization,
	MetricNodeMemoryReservation,
}
var NetworkMetrics = []Metric{
	MetricNetworkRx,
	MetricNetworkRxErrors,
	MetricNetworkRxRate,
	MetricNetworkTx,
	MetricNetworkTxErrors,
	MetricNetworkTxRate,
}

type MetricFamily string

const (
	MetricFamilyCpu        MetricFamily = "cpu"
	MetricFamilyFilesystem              = "filesystem"
	MetricFamilyMemory                  = "memory"
	MetricFamilyNetwork                 = "network"
	MetricFamilyGeneral                 = "general"
)

var MetricFamilies = map[MetricFamily][]Metric{
	MetricFamilyCpu:        CpuMetrics,
	MetricFamilyFilesystem: FilesystemMetrics,
	MetricFamilyMemory:     MemoryMetrics,
	MetricFamilyNetwork:    NetworkMetrics,
}

func MetricFamilyForName(metricName string) MetricFamily {
	for family, metrics := range MetricFamilies {
		for _, metric := range metrics {
			if metricName == metric.Name {
				return family
			}
		}
	}
	return MetricFamilyGeneral
}

var AllMetrics = append(append(append(append(StandardMetrics, AdditionalMetrics...), RateMetrics...), LabeledMetrics...),
	NodeAutoscalingMetrics...)

// Definition of Standard Metrics.
var MetricUptime = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "uptime",
		Description: "Number of milliseconds since the container was started",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsMilliseconds,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return !spec.CreationTime.IsZero()
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		return MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   time.Since(c.Spec.CreationTime).Nanoseconds() / time.Millisecond.Nanoseconds()}
	},
}

var MetricCpuUsage = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/usage",
		Description: "Cumulative CPU usage on all cores",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsNanoseconds,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasCpu
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Cpu.Usage.Total)
		}
		return metricValue
	},
}

var MetricMemoryUsage = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/usage",
		Description: "Memory usage",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasMemory
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricGauge,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Memory.Usage)
		}
		return metricValue
	},
}

var MetricMemoryCache = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/cache",
		Description: "Cache memory",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasMemory
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricGauge,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Memory.Cache)
		}
		return metricValue
	},
}

var MetricMemoryRSS = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/rss",
		Description: "RSS memory",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasMemory
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricGauge,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Memory.RSS)
		}
		return metricValue
	},
}

var MetricMemoryWorkingSet = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/working_set",
		Description: "Working set memory. Working set is the memory being used and not easily dropped by the kernel",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasMemory
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricGauge,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Memory.WorkingSet)
		}
		return metricValue
	},
}

var MetricMemoryPageFaults = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/page_faults",
		Description: "Number of page faults",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasMemory
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Memory.ContainerData.Pgfault)
		}
		return metricValue
	},
}

var MetricMemoryMajorPageFaults = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/major_page_faults",
		Description: "Number of major page faults",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasMemory
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = int64(stat.Memory.ContainerData.Pgmajfault)
		}
		return metricValue
	},
}

var MetricNetworkRx = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/rx",
		Description: "Cumulative number of bytes received over the network",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasNetwork
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var rxBytes int64 = 0
		for _, interfaceStat := range stat.Network.Interfaces {
			rxBytes += int64(interfaceStat.RxBytes)
		}
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = rxBytes
		}
		return metricValue
	},
}

var MetricNetworkRxErrors = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/rx_errors",
		Description: "Cumulative number of errors while receiving over the network",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasNetwork
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var rxErrors int64 = 0
		for _, interfaceStat := range stat.Network.Interfaces {
			rxErrors += int64(interfaceStat.RxErrors)
		}
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = rxErrors
		}
		return metricValue
	},
}

var MetricNetworkTx = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/tx",
		Description: "Cumulative number of bytes sent over the network",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasNetwork
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var txBytes int64 = 0
		for _, interfaceStat := range stat.Network.Interfaces {
			txBytes += int64(interfaceStat.TxBytes)
		}
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = txBytes
		}
		return metricValue
	},
}

var MetricNetworkTxErrors = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/tx_errors",
		Description: "Cumulative number of errors while sending over the network",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
	HasValue: func(spec *cadvisor.ContainerSpec) bool {
		return spec.HasNetwork
	},
	GetValue: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) MetricValue {
		var txErrors int64 = 0
		for _, interfaceStat := range stat.Network.Interfaces {
			txErrors += int64(interfaceStat.TxErrors)
		}
		var metricValue = MetricValue{
			ValueType:  ValueInt64,
			MetricType: MetricCumulative,
			IntValue:   0,
		}
		if !metrics.IsNode(c) {
			metricValue.IntValue = txErrors
		}
		return metricValue
	},
}

// Definition of Additional Metrics.
var MetricCpuRequest = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/request",
		Description: "CPU request (the guaranteed amount of resources) in millicores. This metric is Kubernetes specific.",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
}

var MetricCpuLimit = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/limit",
		Description: "CPU hard limit in millicores.",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
}

var MetricMemoryRequest = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/request",
		Description: "Memory request (the guaranteed amount of resources) in bytes. This metric is Kubernetes specific.",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
}

var MetricMemoryLimit = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/limit",
		Description: "Memory hard limit in bytes.",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
}

// Definition of Rate Metrics.
var MetricCpuUsageRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/usage_rate",
		Description: "CPU usage on all cores in millicores",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
}

var MetricMemoryPageFaultsRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/page_faults_rate",
		Description: "Rate of page faults in counts per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricMemoryMajorPageFaultsRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/major_page_faults_rate",
		Description: "Rate of major page faults in counts per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNetworkRxRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/rx_rate",
		Description: "Rate of bytes received over the network in bytes per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNetworkRxErrorsRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/rx_errors_rate",
		Description: "Rate of errors sending over the network in errors per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNetworkTxRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/tx_rate",
		Description: "Rate of bytes transmitted over the network in bytes per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNetworkTxErrorsRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "network/tx_errors_rate",
		Description: "Rate of errors transmitting over the network in errors per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeCpuCapacity = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/node_capacity",
		Description: "Cpu capacity of a node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeMemoryCapacity = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/node_capacity",
		Description: "Memory capacity of a node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeCpuAllocatable = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/node_allocatable",
		Description: "Cpu allocatable of a node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeMemoryAllocatable = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/node_allocatable",
		Description: "Memory allocatable of a node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeCpuUtilization = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/node_utilization",
		Description: "Cpu utilization as a share of node capacity",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeMemoryUtilization = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/node_utilization",
		Description: "Memory utilization as a share of memory capacity",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeCpuReservation = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "cpu/node_reservation",
		Description: "Share of cpu that is reserved on the node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

var MetricNodeMemoryReservation = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "memory/node_reservation",
		Description: "Share of memory that is reserved on the node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
	},
}

// Labeled metrics

var MetricFilesystemUsage = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "filesystem/usage",
		Description: "Total number of bytes consumed on a filesystem",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		return spec.HasFilesystem
	},
	GetLabeledMetric: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.Filesystem))
		for _, fs := range stat.Filesystem {
			result = append(result, LabeledMetric{
				Name: "filesystem/usage",
				Labels: map[string]string{
					LabelResourceID.Key: fs.Device,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(fs.Usage),
				},
			})
		}
		return result
	},
}

var MetricFilesystemLimit = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "filesystem/limit",
		Description: "The total size of filesystem in bytes",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		return spec.HasFilesystem
	},
	GetLabeledMetric: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.Filesystem))
		for _, fs := range stat.Filesystem {
			result = append(result, LabeledMetric{
				Name: "filesystem/limit",
				Labels: map[string]string{
					LabelResourceID.Key: fs.Device,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(fs.Limit),
				},
			})
		}
		return result
	},
}

var MetricFilesystemAvailable = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "filesystem/available",
		Description: "The number of available bytes remaining in a the filesystem",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
}

var MetricFilesystemInodes = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "filesystem/inodes",
		Description: "Total number of inodes on a filesystem",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		return spec.HasFilesystem
	},
	GetLabeledMetric: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := []LabeledMetric{}
		for _, fs := range stat.Filesystem {
			if fs.HasInodes {
				result = append(result, LabeledMetric{
					Name: "filesystem/inodes",
					Labels: map[string]string{
						LabelResourceID.Key: fs.Device,
					},
					MetricValue: MetricValue{
						ValueType:  ValueInt64,
						MetricType: MetricGauge,
						IntValue:   int64(fs.Inodes),
					},
				})
			}
		}
		return result
	},
}

var MetricFilesystemInodesFree = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "filesystem/inodes_free",
		Description: "Free number of inodes on a filesystem",
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		return spec.HasFilesystem
	},
	GetLabeledMetric: func(c *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := []LabeledMetric{}
		for _, fs := range stat.Filesystem {
			if fs.HasInodes {
				result = append(result, LabeledMetric{
					Name: "filesystem/inodes_free",
					Labels: map[string]string{
						LabelResourceID.Key: fs.Device,
					},
					MetricValue: MetricValue{
						ValueType:  ValueInt64,
						MetricType: MetricGauge,
						IntValue:   int64(fs.InodesFree),
					},
				})
			}
		}
		return result
	},
}

var MetricDiskIORead = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "disk/io_read_bytes",
		Description: "Cumulative number of bytes read over disk",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		return spec.HasDiskIo
	},
	GetLabeledMetric: func(info *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.DiskIo.IoServiceBytes))
		for _, ioServiceBytesPerPartition := range stat.DiskIo.IoServiceBytes {
			resourceIDKey := ioServiceBytesPerPartition.Device
			if resourceIDKey == "" {
				resourceIDKey = fmt.Sprintf("%d:%d", ioServiceBytesPerPartition.Major, ioServiceBytesPerPartition.Minor)
			}

			var value uint64
			if v, exists := ioServiceBytesPerPartition.Stats["Read"]; exists {
				value = v
			}

			result = append(result, LabeledMetric{
				Name: "disk/io_read_bytes",
				Labels: map[string]string{
					LabelResourceID.Key: resourceIDKey,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(value),
				},
			})
		}
		return result
	},
}

var MetricDiskIOWrite = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "disk/io_write_bytes",
		Description: "Cumulative number of bytes write over disk",
		Type:        MetricCumulative,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
		Labels:      metricLabels,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		return spec.HasDiskIo
	},
	GetLabeledMetric: func(info *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.DiskIo.IoServiceBytes))
		for _, ioServiceBytesPerPartition := range stat.DiskIo.IoServiceBytes {
			resourceIDKey := ioServiceBytesPerPartition.Device
			if resourceIDKey == "" {
				resourceIDKey = fmt.Sprintf("%d:%d", ioServiceBytesPerPartition.Major, ioServiceBytesPerPartition.Minor)
			}

			var value uint64
			if v, exists := ioServiceBytesPerPartition.Stats["Write"]; exists {
				value = v
			}

			result = append(result, LabeledMetric{
				Name: "disk/io_write_bytes",
				Labels: map[string]string{
					LabelResourceID.Key: resourceIDKey,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(value),
				},
			})
		}
		return result
	},
}

var MetricDiskIOReadRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "disk/io_read_bytes_rate",
		Description: "Rate of bytes read over disk in bytes per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
		Labels:      metricLabels,
	},
}

var MetricDiskIOWriteRate = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "disk/io_write_bytes_rate",
		Description: "Rate of bytes written over disk in bytes per second",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
		Labels:      metricLabels,
	},
}

var MetricAcceleratorMemoryTotal = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "accelerator/memory_total",
		Description: "Total accelerator memory (in bytes)",
		Labels:      acceleratorLabels,
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		if len(stat.Accelerators) == 0 {
			return false
		}

		return true
	},
	GetLabeledMetric: func(info *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.Accelerators))
		for _, ac := range stat.Accelerators {
			result = append(result, LabeledMetric{
				Name: "accelerator/memory_total",
				Labels: map[string]string{
					LabelAcceleratorMake.Key:  ac.Make,
					LabelAcceleratorModel.Key: ac.Model,
					LabelAcceleratorID.Key:    ac.ID,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(ac.MemoryTotal),
				},
			})
		}
		return result
	},
}

var MetricAcceleratorMemoryUsed = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "accelerator/memory_used",
		Description: "Total accelerator memory allocated (in bytes)",
		Labels:      acceleratorLabels,
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsBytes,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		if len(stat.Accelerators) == 0 {
			return false
		}

		return true
	},
	GetLabeledMetric: func(info *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.Accelerators))
		for _, ac := range stat.Accelerators {
			result = append(result, LabeledMetric{
				Name: "accelerator/memory_used",
				Labels: map[string]string{
					LabelAcceleratorMake.Key:  ac.Make,
					LabelAcceleratorModel.Key: ac.Model,
					LabelAcceleratorID.Key:    ac.ID,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(ac.MemoryUsed),
				},
			})
		}
		return result
	},
}

var MetricAcceleratorDutyCycle = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "accelerator/duty_cycle",
		Description: "Percent of time over the past sample period (10s) during which the accelerator was actively processing",
		Labels:      acceleratorLabels,
		Type:        MetricGauge,
		ValueType:   ValueInt64,
		Units:       UnitsCount,
	},
	HasLabeledMetric: func(spec *cadvisor.ContainerSpec, stat *cadvisor.ContainerStats) bool {
		if len(stat.Accelerators) == 0 {
			return false
		}

		return true
	},
	GetLabeledMetric: func(info *cadvisor.ContainerInfo, stat *cadvisor.ContainerStats) []LabeledMetric {
		result := make([]LabeledMetric, 0, len(stat.Accelerators))
		for _, ac := range stat.Accelerators {
			result = append(result, LabeledMetric{
				Name: "accelerator/duty_cycle",
				Labels: map[string]string{
					LabelAcceleratorMake.Key:  ac.Make,
					LabelAcceleratorModel.Key: ac.Model,
					LabelAcceleratorID.Key:    ac.ID,
				},
				MetricValue: MetricValue{
					ValueType:  ValueInt64,
					MetricType: MetricGauge,
					IntValue:   int64(ac.DutyCycle),
				},
			})
		}
		return result
	},
}

var MetricNodeAcceleratorCapacity = Metric{
	MetricDescriptor: MetricDescriptor{
		Name:        "accelerator/node_capacity",
		Description: "Accelerator capacity of a node",
		Type:        MetricGauge,
		ValueType:   ValueFloat,
		Units:       UnitsCount,
		Labels:      acceleratorCapacityLabels,
	},
}

func IsNodeAutoscalingMetric(name string) bool {
	for _, autoscalingMetric := range NodeAutoscalingMetrics {
		if autoscalingMetric.MetricDescriptor.Name == name {
			return true
		}
	}
	return false
}

type MetricDescriptor struct {
	// The unique name of the metric.
	Name string `json:"name,omitempty"`

	// Description of the metric.
	Description string `json:"description,omitempty"`

	// Descriptor of the labels specific to this metric.
	Labels []LabelDescriptor `json:"labels,omitempty"`

	// Type and value of metric data.
	Type      MetricType `json:"type,omitempty"`
	ValueType ValueType  `json:"value_type,omitempty"`
	Units     UnitsType  `json:"units,omitempty"`
}

// Metric represents a resource usage stat metric.
type Metric struct {
	MetricDescriptor

	// Returns whether this metric is present.
	HasValue func(*cadvisor.ContainerSpec) bool

	// Returns a slice of internal point objects that contain metric values and associated labels.
	GetValue func(*cadvisor.ContainerInfo, *cadvisor.ContainerStats) MetricValue

	// Returns whether this metric is present.
	HasLabeledMetric func(*cadvisor.ContainerSpec, *cadvisor.ContainerStats) bool

	// Returns a slice of internal point objects that contain metric values and associated labels.
	GetLabeledMetric func(*cadvisor.ContainerInfo, *cadvisor.ContainerStats) []LabeledMetric
}
