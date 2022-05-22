package mem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

func TestService(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
	bucket  string
	storage *memoryStorage
}

func (s *Suite) SetupTest() {
	s.bucket = "bucket"
	s.storage = &memoryStorage{
		data: &sync.Map{},
		srv: http.Server{
			Addr: ":" + os.Getenv("PORT"),
		},
	}
}

func (s *Suite) TestDelete() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("abc")
	s.storage.data.Store(fileName, dataUnit{
		bytes: content,
	})

	err := s.storage.Delete(ctx, fileName)
	s.NoError(err)

	_, ok := s.storage.data.Load(fileName)
	s.Equal(ok, false)
}

func (s *Suite) TestDownloadBytes() {
	ctx := context.Background()
	fileName := "fileName"
	content := []byte("abc")

	s.storage.data.Store(fileName, dataUnit{
		bytes: content,
	})

	gotContent, err := s.storage.DownloadBytes(ctx, fileName)
	s.Equal(content, gotContent)
	s.NoError(err)
}

func (s *Suite) TestDownloadByChunksWithNonExistFile() {
	ctx := context.Background()
	fileName := "fileName"

	_, err := s.storage.DownloadByChunks(ctx, fileName)
	s.Error(err)
}

func (s *Suite) TestDownloadByChunksWithExistFile() {
	ctx := context.Background()
	fileName := "fileName"
	data := []byte("abc")
	bytesReader := bytes.NewReader(data)
	bytesReadCloser := io.NopCloser(bytesReader)

	s.storage.data.Store(fileName, dataUnit{
		bytes: data,
	})

	gotContent, err := s.storage.DownloadByChunks(ctx, fileName)
	s.Equal(bytesReadCloser, gotContent)
	s.NoError(err)
}

func (s *Suite) TestUploadBytes() {
	ctx := context.Background()
	fileName := "fileName"
	data := []byte("abc")

	err := s.storage.UploadBytes(ctx, data, fileName)
	s.NoError(err)
}

func (s *Suite) TestUploadByChunks() {
	ctx := context.Background()
	fileName := "fileName"
	data := []byte("abc")
	bytesReader := bytes.NewReader(data)

	err := s.storage.UploadByChunks(ctx, bytesReader, fileName)
	s.NoError(err)

	_, ok := s.storage.data.Load(fileName)
	s.Equal(ok, true)
}

func (s *Suite) TestGenerateGetObjectSignedURL() {
	ctx := context.Background()
	fileName := "fileName"
	data := []byte("abc")

	err := s.storage.UploadBytes(ctx, data, fileName)
	s.NoError(err)
	link, err := s.storage.GenerateGetObjectSignedURL(ctx, fileName, time.Now().Add(time.Minute))
	s.NoError(err)
	fmt.Println(link)
}
