package files

import (
	"context"
	"io"
	"time"
)

type Storage interface {
	PutObject(ctx context.Context, objectKey string, body io.Reader, sizeBytes int64, contentType string) error
	DeleteObject(ctx context.Context, objectKey string) error
	PresignGetObject(ctx context.Context, objectKey string, expiresIn time.Duration) (string, error)
}
