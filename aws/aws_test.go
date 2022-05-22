package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"git.epam.com/epm-gdsp/cloud-uploader-lab/mocks/aws"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
	bucket          string
	awsClient       *AWSBucket
	s3Client        mocks.S3Client
	s3PresignClient mocks.S3PresignClient
}

func (s *Suite) SetupTest() {
	s.s3Client = mocks.S3Client{}
	s.s3PresignClient = mocks.S3PresignClient{}
	s.bucket = "bucket"
	s.awsClient = &AWSBucket{
		client:   &s.s3Client,
		bucket:   s.bucket,
		psClient: &s.s3PresignClient,
	}
}

func (s *Suite) TestUploadBytes() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("abc")
	e := fmt.Errorf("error")
	putObjectInput := s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &fileName,
		Body:   bytes.NewReader(content),
	}
	tests := map[string]struct {
		err, expectedErr error
	}{
		"without_error": {
			err:         nil,
			expectedErr: nil,
		},
		"error": {
			err:         e,
			expectedErr: fmt.Errorf("error"),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.s3Client.On("PutObject", ctx, &putObjectInput).Once().Return(nil, test.err)
			err := s.awsClient.UploadBytes(ctx, content, fileName)
			s.Equal(test.expectedErr, errors.Unwrap(errors.Unwrap(err)))
		})
	}
}

func (s *Suite) TestDownloadBytes() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("abc")
	e := fmt.Errorf("error")
	getObjectInput := s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &fileName,
	}
	getObjectOutput := &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(content)),
	}
	tests := map[string]struct {
		err, expectedErr error
		expectedContent  []byte
		output           *s3.GetObjectOutput
	}{
		"without_error": {
			err:             nil,
			expectedErr:     nil,
			expectedContent: content,
			output:          getObjectOutput,
		},
		"error": {
			err:             e,
			expectedErr:     fmt.Errorf("error"),
			expectedContent: nil,
			output:          nil,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.s3Client.On("GetObject", ctx, &getObjectInput).Once().Return(test.output, test.err)
			gotContent, err := s.awsClient.DownloadBytes(ctx, fileName)
			s.Equal(test.expectedContent, gotContent)
			s.Equal(test.expectedErr, errors.Unwrap(errors.Unwrap(err)))
		})
	}
}

func (s *Suite) TestDelete() {
	ctx := context.Background()
	fileName := "fileName"
	e := fmt.Errorf("error")
	deleteObjectInput := s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &fileName,
	}
	tests := map[string]struct {
		err, expectedErr error
	}{
		"without_error": {
			err:         nil,
			expectedErr: nil,
		},
		"error": {
			err:         e,
			expectedErr: fmt.Errorf("error"),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.s3Client.On("DeleteObject", ctx, &deleteObjectInput).Once().Return(nil, test.err)
			err := s.awsClient.Delete(ctx, fileName)
			s.Equal(test.expectedErr, errors.Unwrap(err))
		})
	}
}

func (s *Suite) TestGenerateObjectSignURL() {
	ctx := context.Background()
	fileName := "fileName"
	url := "https://example.com"
	err := errors.New("some error")
	ttl := time.Now().Add(time.Hour)

	presignGetObjectInput := s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &fileName,
	}
	presignPutObjectInput := s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &fileName,
	}

	type Tests struct {
		psHTTPRequest    *v4.PresignedHTTPRequest
		err, expectedErr error
		expectedURL      string
	}
	testError := Tests{
		psHTTPRequest: nil,
		err:           err,
		expectedURL:   "",
		expectedErr:   err,
	}
	testWithoutError := Tests{
		psHTTPRequest: &v4.PresignedHTTPRequest{URL: url},
		err:           nil,
		expectedURL:   url,
		expectedErr:   nil,
	}

	tests := map[string]Tests{
		"PresignGetObject_without_error": testWithoutError,
		"PresignGetObject_error":         testError,
		"PresignPutObject_without_error": testWithoutError,
		"PresignPutObject_error":         testError,
	}

	for name, test := range tests {
		s.Run(name, func() {
			method := strings.Split(name, "_")[0]
			var gotUrl string
			var err error
			if method == "PresignGetObject" {
				s.s3PresignClient.On(method, ctx, &presignGetObjectInput, mock.Anything).Once().Return(test.psHTTPRequest, test.err)
				gotUrl, err = s.awsClient.GenerateGetObjectSignedURL(ctx, fileName, ttl)
			} else {
				s.s3PresignClient.On(method, ctx, &presignPutObjectInput, mock.Anything).Once().Return(test.psHTTPRequest, test.err)
				gotUrl, err = s.awsClient.GeneratePutObjectSignedURL(ctx, fileName, ttl)
			}
			s.Equal(test.expectedURL, gotUrl)
			s.Equal(test.expectedErr, errors.Unwrap(err))
		})
	}
}
