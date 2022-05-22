package azure

import (
	"bytes"
	"context"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/mocks/azure"
	"github.com/stretchr/testify/suite"
	"io"
	"io/ioutil"
	"testing"
	"time"
)

type reader struct {
	io.ReadCloser
}

func TestService(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
	bucket  string
	azure   *bucketAzure
	adapter mocks.AdapterInterface
}

func (s *Suite) SetupTest() {
	s.adapter = mocks.AdapterInterface{}
	s.bucket = "bucket"
	s.azure = &bucketAzure{
		bucketName: s.bucket,
		newAdapter: func(ctx context.Context) (adapterInterface, error) {
			return &s.adapter, nil
		},
	}
}

func (s *Suite) TestUploadBytesSuccess() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("UsefulInfo")

	s.adapter.On("Upload", content, s.bucket, fileName).Once().Return(nil)
	err := s.azure.UploadBytes(ctx, content, fileName)
	s.NoError(err)
}

func (s *Suite) TestUploadChunksSuccess() {
	ctx := context.Background()
	fileName := "fileName"
	fileAsReader := reader{}

	s.adapter.On("UploadChunks", fileAsReader, s.bucket, fileName).Once().Return(nil)
	err := s.azure.UploadByChunks(ctx, fileAsReader, fileName)
	s.NoError(err)
}

func (s *Suite) TestDeleteSuccess() {
	ctx := context.Background()
	fileName := "fileName"

	s.adapter.On("Delete", s.bucket, fileName).Once().Return(nil)
	err := s.azure.Delete(ctx, fileName)
	s.NoError(err)
}

func (s *Suite) TestGenerateGetObjectSignedURLSuccess() {
	ctx := context.Background()
	fileName := "fileName"
	ttl := time.Time{}

	s.adapter.On("GenerateSignedURL", s.bucket, fileName, ttl).Once().Return("", nil)
	url, _ := s.azure.GenerateGetObjectSignedURL(ctx, fileName, ttl)
	s.Equal(url, "")
}

func (s *Suite) TestDownloadBytesSuccess() {
	ctx := context.Background()
	fileName := "fileName"
	file := []byte("usefulInfo")

	s.adapter.On("DownloadBytes", s.bucket, fileName).Once().Return(ioutil.NopCloser(bytes.NewReader(file)), nil) // reader is not a reader interface

	arr, err := s.azure.DownloadBytes(ctx, fileName)
	s.Equal(arr, file)
	s.NoError(err)
}

func (s *Suite) TestDownloadByChunksSuccess() {
	ctx := context.Background()
	fileName := "fileName"
	file := []byte("usefulInfo")

	s.adapter.On("DownloadBytes", s.bucket, fileName).Once().Return(ioutil.NopCloser(bytes.NewReader(file)), nil) // reader is not a reader interface

	arr, err := s.azure.DownloadByChunks(ctx, fileName)
	s.Equal(arr, ioutil.NopCloser(bytes.NewReader(file)))
	s.NoError(err)
}
