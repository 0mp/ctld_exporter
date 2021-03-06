// SPDX-Licence-Identifier: Apache-2.0
package main

import (
	"github.com/Gandi/ctld_exporter/ctlstats"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"strconv"
)

var (
	ioBytesDesc = prometheus.NewDesc(
		"iscsi_target_bytes",
		"Number of bytes",
		[]string{"type", "target", "lun"}, nil,
	)
	ioOperationsDesc = prometheus.NewDesc(
		"iscsi_target_operations",
		"Number of operations",
		[]string{"type", "target", "lun"}, nil,
	)
	ioDmasDesc = prometheus.NewDesc(
		"iscsi_target_dmas",
		"Number of DMA",
		[]string{"type", "target", "lun"}, nil,
	)
	initiatorsNumberDesc = prometheus.NewDesc(
		"iscsi_target_initiators",
		"Number of initiators connected to target",
		[]string{"target"}, nil,
	)
)

type iscsiCollector struct {
}

func (ic iscsiCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(ic, ch)
}
func (ic iscsiCollector) Collect(ch chan<- prometheus.Metric) {
	dataByLun := ctlstats.GetStats()
	targets := ctlstats.GetTargets()
	for lun, data := range dataByLun {
		stringlun := strconv.FormatUint(uint64(targets.GetLunId(uint(lun))), 10)
		target := targets.GetLunTarget(uint(lun)).Name
		ch <- prometheus.MustNewConstMetric(
			ioBytesDesc,
			prometheus.CounterValue,
			float64(data.Bytes[ctlstats.CTL_STATS_NO_IO]),
			"NO IO", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioBytesDesc,
			prometheus.CounterValue,
			float64(data.Bytes[ctlstats.CTL_STATS_READ]),
			"READ", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioBytesDesc,
			prometheus.CounterValue,
			float64(data.Bytes[ctlstats.CTL_STATS_WRITE]),
			"WRITE", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioOperationsDesc,
			prometheus.CounterValue,
			float64(data.Operations[ctlstats.CTL_STATS_NO_IO]),
			"NO IO", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioOperationsDesc,
			prometheus.CounterValue,
			float64(data.Operations[ctlstats.CTL_STATS_READ]),
			"READ", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioOperationsDesc,
			prometheus.CounterValue,
			float64(data.Operations[ctlstats.CTL_STATS_WRITE]),
			"WRITE", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioDmasDesc,
			prometheus.CounterValue,
			float64(data.Dmas[ctlstats.CTL_STATS_NO_IO]),
			"NO IO", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioDmasDesc,
			prometheus.CounterValue,
			float64(data.Dmas[ctlstats.CTL_STATS_READ]),
			"READ", target, stringlun)
		ch <- prometheus.MustNewConstMetric(
			ioDmasDesc,
			prometheus.CounterValue,
			float64(data.Dmas[ctlstats.CTL_STATS_WRITE]),
			"WRITE", target, stringlun)
	}
	for _, target := range targets.Targets {
		if target.Name == "" {
			continue
		}
		ch <- prometheus.MustNewConstMetric(
			initiatorsNumberDesc,
			prometheus.GaugeValue,
			float64(len(target.Initiators)),
			target.Name)
	}
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9572").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("ctld_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	reg := prometheus.NewPedanticRegistry()
	ic := iscsiCollector{}
	reg.MustRegister(
		ic,
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)
	http.Handle(*metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>ctld Exporter</title></head>
			<body>
			<h1>ctld Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
		if err != nil {
			log.Errorln(err)
		}
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
