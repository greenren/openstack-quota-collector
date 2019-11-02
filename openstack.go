package main

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	bs "github.com/gophercloud/gophercloud/openstack/blockstorage/extensions/quotasets"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/quotasets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type usage struct {
	CPULimit     int
	CPUUsed      int
	RAMUsed      int
	RAMLimit     int
	VolumesUsed  int
	VolumesLimit int
}

type metrics struct {
	CPUUsed     float64
	RAMUsed     float64
	VolumesUsed float64
}

var (
	gaugeCPU = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openstack",
			Name:      "cpu_percentage_inuse",
			Help:      "percentage of cores in use",
		})

	gaugeRAM = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openstack",
			Name:      "ram_percentage_inuse",
			Help:      "percentage of ram in use",
		})

	gaugeVolumes = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "openstack",
			Name:      "volumes_percentage_inuse",
			Help:      "percentage of volumes in use",
		})
)

func usagePercentage(usage, limit int) float64 {

	var percentage float64
	percentage = float64((usage * 100) / limit)
	return percentage

}

func metricsHandler() {

	go func() {

		opts, err := openstack.AuthOptionsFromEnv()

		if err != nil {
			log.WithError(err)
		}

		provider, err := openstack.AuthenticatedClient(opts)

		if err != nil {
			log.WithError(err)
		}

		for {

			client, _ := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{})
			quotaset, _ := quotasets.GetDetail(client, os.Getenv("OS_PROJECT_ID")).Extract()

			clientStorage, _ := openstack.NewBlockStorageV3(provider, gophercloud.EndpointOpts{})
			quotasetStorage, _ := bs.GetUsage(clientStorage, os.Getenv("OS_PROJECT_ID")).Extract()

			log.Info(quotasetStorage)

			usage := new(usage)

			usage.CPULimit = quotaset.Cores.Limit
			usage.CPUUsed = quotaset.Cores.InUse
			usage.RAMLimit = quotaset.RAM.Limit
			usage.RAMUsed = quotaset.RAM.InUse
			usage.VolumesUsed = quotasetStorage.Volumes.InUse
			usage.VolumesLimit = quotasetStorage.Volumes.Limit

			metric := new(metrics)
			metric.CPUUsed = usagePercentage(usage.CPUUsed, usage.CPULimit)
			metric.RAMUsed = usagePercentage(usage.RAMUsed, usage.RAMLimit)
			metric.VolumesUsed = usagePercentage(usage.VolumesUsed, usage.VolumesLimit)

			gaugeCPU.Set(metric.CPUUsed)
			gaugeRAM.Set(metric.RAMUsed)
			gaugeVolumes.Set(metric.VolumesUsed)

			time.Sleep(10 * time.Second)
		}

	}()

}

func main() {

	metricsHandler()

	prometheus.MustRegister(gaugeCPU)
	prometheus.MustRegister(gaugeRAM)
	prometheus.MustRegister(gaugeVolumes)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":9080", nil))

}
