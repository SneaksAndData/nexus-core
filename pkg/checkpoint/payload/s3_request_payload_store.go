package payload

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"k8s.io/klog/v2"
)

const s3UrlRegex = "s3a://([^/]+)/?(.*)"

type S3RequestPayloadStore struct {
	client               *s3.Client
	signer               *s3.PresignClient
	logger               klog.Logger
	uploader             *manager.Uploader
	payloadStoragePrefix string
}

type S3Path struct {
	Bucket *string
	Key    *string
}

func NewS3Path(bucket string, key string) *S3Path {
	return &S3Path{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
}

// NewS3PayloadStore initializes S3 client for the S3 payload store. Providing empty values for credentialsProvider, s3endpoint and s3region will result in using default SDK credential flow.
func NewS3PayloadStore(ctx context.Context, logger klog.Logger, credentialsProvider aws.CredentialsProvider, s3endpoint string, s3region string, payloadStoragePath string) *S3RequestPayloadStore {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.V(0).Error(err, "error when reading S3 configuration")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.UseAccelerate = false
		o.HTTPSignerV4 = v4.NewSigner()
		o.AppID = "nexus"
		o.Credentials = credentialsProvider
		o.BaseEndpoint = &s3endpoint
		o.Region = s3region
	})
	return &S3RequestPayloadStore{
		client:               client,
		signer:               s3.NewPresignClient(client),
		uploader:             manager.NewUploader(client), // using defaults - add support for tuning if needed at some point
		logger:               logger,
		payloadStoragePrefix: payloadStoragePath,
	}
}

func parsePath(blobPath string) *S3Path {
	r, _ := regexp.Compile(s3UrlRegex)
	matches := r.FindStringSubmatch(blobPath)
	return NewS3Path(matches[1], matches[2])
}

func (store *S3RequestPayloadStore) getStoragePath(requestId string, templateName string) string {
	return fmt.Sprintf("%s/%s/%s",
		store.payloadStoragePrefix,
		fmt.Sprintf("algorithm=%s", templateName),
		requestId)
}

func (store *S3RequestPayloadStore) Persist(ctx context.Context, payload string, requestId string, templateName string) error {
	payloadPath := store.getStoragePath(requestId, templateName)
	s3Path := parsePath(payloadPath)

	result, err := store.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: s3Path.Bucket,
		Key:    s3Path.Key,
		Body:   strings.NewReader(payload),
	})

	if err != nil {
		store.logger.V(0).Error(err, "error when persisting payload into S3")
		return err
	}
	store.logger.V(4).Info("successfully persisted algorithm payload", "payloadPath", payloadPath, "etag", *result.ETag)

	return nil
}

func (store *S3RequestPayloadStore) GenerateUrl(ctx context.Context, requestId string, templateName string, validFor time.Duration) (string, error) {
	payloadPath := store.getStoragePath(requestId, templateName)
	s3Path := parsePath(payloadPath)

	result, err := store.signer.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: s3Path.Bucket,
		Key:    s3Path.Key,
	},
		s3.WithPresignExpires(validFor))

	if err != nil {
		store.logger.V(0).Error(err, "error when generating payload URI")
		return "", err
	}

	return result.URL, nil
}
