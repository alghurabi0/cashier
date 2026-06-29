package storage

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Storage provides methods to upload files to Cloudflare R2 (S3-compatible).
type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string // e.g. "https://pub-xxx.r2.dev" or custom domain
}

// NewR2Storage creates a new R2Storage client from environment variables.
// Required env vars: R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME, R2_PUBLIC_URL
func NewR2Storage() (*R2Storage, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKeyID := os.Getenv("R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	// If not configured, return nil (uploads will fail gracefully)
	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &R2Storage{
		client:    client,
		bucket:    bucketName,
		publicURL: publicURL,
	}, nil
}

// Upload stores a file in R2 and returns the public URL.
func (s *R2Storage) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("R2 storage is not configured (set R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME env vars)")
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	publicURL := s.publicURL + "/" + key
	return publicURL, nil
}

// Get retrieves a file from R2 by key. Returns the body, content type, and any error.
// Caller must close the returned ReadCloser.
func (s *R2Storage) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	if s == nil {
		return nil, "", fmt.Errorf("R2 storage is not configured")
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object from R2: %w", err)
	}

	contentType := ""
	if result.ContentType != nil {
		contentType = *result.ContentType
	}

	return result.Body, contentType, nil
}

// PublicURL returns the configured public URL prefix for R2 objects.
func (s *R2Storage) PublicURL() string {
	if s == nil {
		return ""
	}
	return s.publicURL
}

// IsConfigured returns true if R2 storage is properly configured.
func (s *R2Storage) IsConfigured() bool {
	return s != nil && s.client != nil
}
