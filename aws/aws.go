package aws

import (
	"bytes"
	"context"
	"fmt"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/bucket"
	"io"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type s3PresignClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type AWSBucket struct {
	client   s3Client
	bucket   string
	psClient s3PresignClient
}

var _ bucket.Bucket = (*AWSBucket)(nil)

// OpenBucket uses default config (https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk)
func OpenBucket(ctx context.Context, bucket string) (*AWSBucket, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(cfg)
	listBuckets, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	m := 0
	for _, b := range listBuckets.Buckets {
		if *b.Name == bucket {
			m++
		}
	}
	if m != 0 {
		return &AWSBucket{
			client:   s3Client,
			bucket:   bucket,
			psClient: s3.NewPresignClient(s3Client),
		}, nil
	}
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: &bucket,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraintEuCentral1,
		},
	})
	if err != nil {
		return nil, err
	}
	return &AWSBucket{
		client:   s3Client,
		bucket:   bucket,
		psClient: s3.NewPresignClient(s3Client),
	}, nil
}

func (c *AWSBucket) UploadByChunks(ctx context.Context, content io.Reader, filename string) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    &filename,
		Body:   content,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (c *AWSBucket) UploadBytes(ctx context.Context, fileAsBytes []byte, objName string) error {
	err := c.UploadByChunks(ctx, bytes.NewReader(fileAsBytes), objName)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (c *AWSBucket) DownloadByChunks(ctx context.Context, filename string) (io.ReadCloser, error) {
	res, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &filename,
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return res.Body, nil
}

func (c *AWSBucket) DownloadBytes(ctx context.Context, objName string) ([]byte, error) {
	rc, err := c.DownloadByChunks(ctx, objName)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return b, nil
}

func (c *AWSBucket) Delete(ctx context.Context, filename string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &c.bucket,
		Key:    &filename,
	})
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (c *AWSBucket) GenerateGetObjectSignedURL(ctx context.Context, filename string, ttl time.Time) (string, error) {
	presignedHTTPRequest, err := c.psClient.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: &c.bucket,
			Key:    &filename,
		},
		s3.WithPresignExpires(time.Until(ttl)),
	)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return presignedHTTPRequest.URL, nil
}

func (c *AWSBucket) GeneratePutObjectSignedURL(ctx context.Context, filename string, ttl time.Time) (string, error) {
	presignedHTTPRequest, err := c.psClient.PresignPutObject(ctx,
		&s3.PutObjectInput{
			Bucket: &c.bucket,
			Key:    &filename,
		},
		s3.WithPresignExpires(time.Until(ttl)),
	)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return presignedHTTPRequest.URL, nil
}
