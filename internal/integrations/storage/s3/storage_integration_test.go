package s3storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
)

func TestClientPutObjectPresignAndDeleteIntegration(t *testing.T) {
	cfg, err := platformconfig.Load()
	if err != nil {
		t.Skipf("integration config unavailable: %v", err)
	}

	client, err := New(cfg.Storage)
	if err != nil {
		t.Skipf("storage config unavailable: %v", err)
	}

	ctx := context.Background()
	objectKey := fmt.Sprintf("integration-tests/%d/sample.mp3", time.Now().UnixNano())
	payload := []byte("coralhub-stage5")

	if err := client.PutObject(ctx, objectKey, bytes.NewReader(payload), int64(len(payload)), "audio/mpeg"); err != nil {
		t.Skipf("minio unavailable for integration test: %v", err)
	}
	t.Cleanup(func() {
		_ = client.DeleteObject(context.Background(), objectKey)
	})

	url, err := client.PresignGetObject(ctx, objectKey, time.Minute)
	if err != nil {
		t.Fatalf("PresignGetObject() error = %v", err)
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	response, err := httpClient.Get(url)
	if err != nil {
		t.Fatalf("GET presigned url error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("response.StatusCode = %d, want %d", response.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	if string(body) != string(payload) {
		t.Fatalf("body = %q, want %q", body, payload)
	}

	if err := client.DeleteObject(ctx, objectKey); err != nil {
		t.Fatalf("DeleteObject() error = %v", err)
	}
}
