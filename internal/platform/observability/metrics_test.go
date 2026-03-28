package observability

import (
	"strings"
	"testing"
	"time"
)

func TestRegistryRenderIncludesObservedMetrics(t *testing.T) {
	registry := NewRegistry()
	registry.ObserveHTTPRequest("GET", "/healthz", 200, 25*time.Millisecond)
	registry.IncrementWorkerPoll()
	registry.AddWorkerProcessed(3)
	registry.IncrementNotificationDelivery("sent")
	registry.IncrementStorageUploadFailure()
	registry.AddNotificationCleanupDeleted(2)

	rendered := registry.render()

	for _, expected := range []string{
		`coralhub_http_requests_total{method="GET",path="/healthz",status="200"} 1`,
		"coralhub_worker_polls_total 1",
		"coralhub_worker_processed_total 3",
		`coralhub_notification_delivery_total{result="sent"} 1`,
		"coralhub_storage_upload_failures_total 1",
		"coralhub_notification_cleanup_deleted_total 2",
	} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("render() missing %q in:\n%s", expected, rendered)
		}
	}
}
