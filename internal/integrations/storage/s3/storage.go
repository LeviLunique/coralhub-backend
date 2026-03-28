package s3storage

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	platformconfig "github.com/LeviLunique/coralhub-backend/internal/platform/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
}

func New(cfg platformconfig.StorageConfig) (*Client, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, errors.New("storage bucket is required")
	}

	if strings.TrimSpace(cfg.Region) == "" {
		return nil, errors.New("storage region is required")
	}

	awsConfig := aws.Config{
		Region:      cfg.Region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	}

	options := func(options *s3.Options) {
		endpointURL := cfg.EndpointURL()
		if endpointURL != "" {
			options.BaseEndpoint = aws.String(endpointURL)
			options.UsePathStyle = !strings.Contains(endpointURL, "amazonaws.com")
		}
	}

	client := s3.NewFromConfig(awsConfig, options)

	return &Client{
		bucket:  cfg.Bucket,
		client:  client,
		presign: s3.NewPresignClient(client),
	}, nil
}

func (c *Client) PutObject(ctx context.Context, objectKey string, body io.Reader, sizeBytes int64, contentType string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(objectKey),
		Body:          body,
		ContentLength: aws.Int64(sizeBytes),
		ContentType:   aws.String(contentType),
	})

	return err
}

func (c *Client) DeleteObject(ctx context.Context, objectKey string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	})
	return err
}

func (c *Client) PresignGetObject(ctx context.Context, objectKey string, expiresIn time.Duration) (string, error) {
	result, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(objectKey),
	}, func(options *s3.PresignOptions) {
		options.Expires = expiresIn
	})
	if err != nil {
		return "", err
	}

	return result.URL, nil
}
