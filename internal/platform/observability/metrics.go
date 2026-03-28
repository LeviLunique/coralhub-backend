package observability

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Registry struct {
	mu                           sync.Mutex
	httpRequests                 map[string]uint64
	httpRequestDurationCount     map[string]uint64
	httpRequestDurationSumMillis map[string]float64
	workerPollsTotal             uint64
	workerProcessedTotal         uint64
	notificationDeliveryTotal    map[string]uint64
	notificationCleanupDeleted   uint64
	storageUploadFailuresTotal   uint64
}

func NewRegistry() *Registry {
	return &Registry{
		httpRequests:                 make(map[string]uint64),
		httpRequestDurationCount:     make(map[string]uint64),
		httpRequestDurationSumMillis: make(map[string]float64),
		notificationDeliveryTotal:    make(map[string]uint64),
	}
}

var defaultRegistry = NewRegistry()

func DefaultMetrics() *Registry {
	return defaultRegistry
}

func (r *Registry) ObserveHTTPRequest(method string, path string, status int, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := fmt.Sprintf("method=%s,path=%s,status=%d", sanitizeLabel(method), sanitizeLabel(path), status)
	r.httpRequests[key]++
	r.httpRequestDurationCount[key]++
	r.httpRequestDurationSumMillis[key] += float64(duration) / float64(time.Millisecond)
}

func (r *Registry) IncrementWorkerPoll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.workerPollsTotal++
}

func (r *Registry) AddWorkerProcessed(count int) {
	if count <= 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.workerProcessedTotal += uint64(count)
}

func (r *Registry) IncrementNotificationDelivery(result string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.notificationDeliveryTotal[sanitizeLabel(result)]++
}

func (r *Registry) AddNotificationCleanupDeleted(count int64) {
	if count <= 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.notificationCleanupDeleted += uint64(count)
}

func (r *Registry) IncrementStorageUploadFailure() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.storageUploadFailuresTotal++
}

func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(r.render()))
	})
}

func (r *Registry) render() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lines []string
	lines = append(lines,
		"# TYPE coralhub_http_requests_total counter",
		"# TYPE coralhub_http_request_duration_milliseconds_count counter",
		"# TYPE coralhub_http_request_duration_milliseconds_sum counter",
	)

	httpKeys := sortedKeys(r.httpRequests)
	for _, key := range httpKeys {
		labels := toPrometheusLabels(key)
		lines = append(lines,
			fmt.Sprintf("coralhub_http_requests_total{%s} %d", labels, r.httpRequests[key]),
			fmt.Sprintf("coralhub_http_request_duration_milliseconds_count{%s} %d", labels, r.httpRequestDurationCount[key]),
			fmt.Sprintf("coralhub_http_request_duration_milliseconds_sum{%s} %g", labels, r.httpRequestDurationSumMillis[key]),
		)
	}

	lines = append(lines,
		"# TYPE coralhub_worker_polls_total counter",
		fmt.Sprintf("coralhub_worker_polls_total %d", r.workerPollsTotal),
		"# TYPE coralhub_worker_processed_total counter",
		fmt.Sprintf("coralhub_worker_processed_total %d", r.workerProcessedTotal),
		"# TYPE coralhub_notification_cleanup_deleted_total counter",
		fmt.Sprintf("coralhub_notification_cleanup_deleted_total %d", r.notificationCleanupDeleted),
		"# TYPE coralhub_storage_upload_failures_total counter",
		fmt.Sprintf("coralhub_storage_upload_failures_total %d", r.storageUploadFailuresTotal),
		"# TYPE coralhub_notification_delivery_total counter",
	)

	deliveryKeys := sortedKeys(r.notificationDeliveryTotal)
	for _, key := range deliveryKeys {
		lines = append(lines, fmt.Sprintf(`coralhub_notification_delivery_total{result="%s"} %d`, key, r.notificationDeliveryTotal[key]))
	}

	return strings.Join(lines, "\n") + "\n"
}

func sortedKeys[T any](items map[string]T) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sanitizeLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}
	return strings.ReplaceAll(trimmed, `"`, `'`)
}

func toPrometheusLabels(value string) string {
	parts := strings.Split(value, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		key, raw, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		labels = append(labels, fmt.Sprintf(`%s="%s"`, key, raw))
	}
	return strings.Join(labels, ",")
}
