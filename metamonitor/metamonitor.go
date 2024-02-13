package metamonitor

import "github.com/prometheus/client_golang/prometheus"

const (
	SUBSYSTEM      = "patient_service"
	NAMESPACE      = "intuitive"
	LABEL_WILDCARD = "*"
)

// Meta monitoring counters
type MetaMonitorMetricType struct {
	dBRequestLatency        *prometheus.SummaryVec
	jobLatency              *prometheus.SummaryVec
	granularLabels          bool
	metaMonitorDecoratorUrl string
}

var META_MONITOR *MetaMonitorMetricType

func init() {

	META_MONITOR = &MetaMonitorMetricType{
		dBRequestLatency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "request_latency_seconds",
			Namespace: NAMESPACE,
			Subsystem: SUBSYSTEM,
			Help:      "Latency measure for app manager db requests",
		},
			[]string{"api"}),
		jobLatency: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:      "app_manager_job_completion_latency_seconds",
			Namespace: NAMESPACE,
			Subsystem: SUBSYSTEM,
			Help:      "Initial Completion time Latency measure for app manager jobs",
		},
			[]string{"job_type"}),
		granularLabels: false,
	}

	prometheus.MustRegister(META_MONITOR.dBRequestLatency)
	prometheus.MustRegister(META_MONITOR.jobLatency)

}

// Setter

func (ms *MetaMonitorMetricType) SetGranularLabels(enable bool) {
	ms.granularLabels = enable
}

func (ms *MetaMonitorMetricType) SetMetaMonitorDecoratorUrl(url string) {
	ms.metaMonitorDecoratorUrl = url
}

func (ms *MetaMonitorMetricType) GetDbRequestLatencyVector() *prometheus.SummaryVec {
	return ms.dBRequestLatency
}

func (ms *MetaMonitorMetricType) GetJobLatencyVector() *prometheus.SummaryVec {
	return ms.jobLatency
}

// Getter

func (ms *MetaMonitorMetricType) GetMetaMonitorDecoratorUrl() string {
	if ms != nil {
		return ms.metaMonitorDecoratorUrl
	}
	return ""
}

func (ms *MetaMonitorMetricType) IsGranularLabelsEnabled() bool {
	return ms.granularLabels
}

// If granular label tracking is on , return actual value, else collapse
// all labels into generic wild card
func GetGranularLabel(value string) string {
	if META_MONITOR.granularLabels {
		return value
	}
	return LABEL_WILDCARD
}
