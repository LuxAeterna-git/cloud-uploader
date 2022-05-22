package gcp

import (
	"bytes"
	"cloud-uploader-lab/mocks/gcp"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/storage"

	"github.com/stretchr/testify/suite"
)

func NopWriteCloser(r io.Writer) io.WriteCloser {
	return nopWriteCloser{r}
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error { return nil }

func TestService(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
	bucket  string
	gcp     *bucketGCP
	adapter mocks.AdapterInterface
}

func (s *Suite) SetupTest() {
	s.adapter = mocks.AdapterInterface{}
	s.bucket = "bucket"
	s.gcp = &bucketGCP{
		bucketName: s.bucket,
		newAdapter: func(ctx context.Context) (adapterInterface, error) {
			return &s.adapter, nil
		},
	}
}

func NopCloser(r io.Writer) io.WriteCloser {
	return nopCloser{r}
}

type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

func (s *Suite) TestUploadBytes() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("abc")
	buf := &bytes.Buffer{}
	s.adapter.On("NewWriter", fileName, s.bucket).Once().
		Return(NopCloser(buf), nil)
	s.adapter.On("Close").Once().Return(nil)
	err := s.gcp.UploadBytes(ctx, content, fileName)
	s.Equal(content, buf.Bytes())
	s.NoError(err)

}

func (s *Suite) TestDownloadBytes() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("abc")
	s.adapter.On("NewReader", fileName, s.bucket).Once().
		Return(io.NopCloser(bytes.NewReader(content)), nil)
	s.adapter.On("Close").Once().Return(nil)
	gotContent, err := s.gcp.DownloadBytes(ctx, fileName)
	s.Equal(content, gotContent)
	s.NoError(err)
}

func (s *Suite) TestDelete() {
	ctx := context.Background()
	fileName := "fileName"
	s.adapter.On("Delete", fileName, s.bucket).Once().
		Return(nil)
	s.adapter.On("Close").Once().Return(nil)
	err := s.gcp.Delete(ctx, fileName)

	s.NoError(err)
}

func (s *Suite) TestGenerateGetObjectSignedURL() {
	ctx := context.Background()
	fileName := "fileName"
	url := "testurl"
	ttl := time.Now().Add(15 * time.Minute)
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         http.MethodGet,
		GoogleAccessID: "conf.Email",
		PrivateKey:     []byte("conf.PrivateKey"),
		Expires:        ttl,
	}
	s.adapter.On("OptsGen", ttl).Once().
		Return(opts, nil)
	s.adapter.On("SignedURL", s.bucket, fileName, opts).Once().
		Return(url, nil)
	s.adapter.On("Close").Once().Return(nil)
	returnedURL, err := s.gcp.GenerateGetObjectSignedURL(ctx, fileName, ttl)
	s.Equal(url, returnedURL)
	s.NoError(err)
}

func (s *Suite) TestDownloadByChunks() {
	ctx := context.Background()
	fileName := "fileName"
	stringReader := strings.NewReader("abc")
	stringReadCloser := io.NopCloser(stringReader)
	s.adapter.On("NewReader", fileName, s.bucket).Once().
		Return(stringReadCloser, nil)
	s.adapter.On("Close").Once().Return(nil)
	gotContent, err := s.gcp.DownloadByChunks(ctx, fileName)
	s.Equal(stringReadCloser, gotContent)
	s.NoError(err)
}

func (s *Suite) TestUploadByChunks() {
	ctx := context.Background()
	fileName := "fileName"
	content := strings.NewReader("abc")
	buf := &bytes.Buffer{}
	s.adapter.On("NewWriter", fileName, s.bucket).Once().
		Return(NopWriteCloser(buf), nil)
	s.adapter.On("Close").Once().Return(nil)
	err := s.gcp.UploadByChunks(ctx, content, fileName)
	s.NoError(err)
}
