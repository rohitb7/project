package core

import "github.com/prometheus/client_golang/prometheus"

const (
	SUBSYSTEM       = "blob_service"
	NAMESPACE       = "intuitive"
	MetaMonitorOK   = "ok"
	MetaMonitorFAIL = "fail"
)

// Meta monitoring counters
type MetaMonitorMetricType struct {
	storageAPIRequestLatency  *prometheus.SummaryVec
	storageAPICounter         *prometheus.CounterVec
	storageObjectCountCounter *prometheus.CounterVec
}

var StorageMetaMonitor *MetaMonitorMetricType

func init() {

	StorageMetaMonitor = &MetaMonitorMetricType{
		storageAPICounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "s3_api_count",
			Namespace: NAMESPACE,
			Subsystem: SUBSYSTEM,
			Help:      "Counter for s3 api calls",
		},
			// Labels to use for counter. This should match with params in withLabels
			[]string{"api", "operation"}),
	}
	prometheus.MustRegister(StorageMetaMonitor.storageAPICounter)

}

func (ms *MetaMonitorMetricType) GetAPICounter() *prometheus.CounterVec {
	return ms.storageAPICounter
}
