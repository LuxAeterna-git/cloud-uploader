package mem

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const chunkSize = 32
const pattern = "/files"
const urlValue = "filename"
const HostName = "HOSTNAME_MEM_SRV"
const Port = "PORT_MEM_SRV"

type ErrNoSetEnvVars struct{}

func (e ErrNoSetEnvVars) Error() string {
	return "Necessary environment variables are not set"
}

type ErrNoSuchObject struct{}

func (e ErrNoSuchObject) Error() string {
	return "No file exist with such name"
}

type ErrTypeAssertion struct{}

func (e ErrTypeAssertion) Error() string {
	return "type assertion error"
}

type dataUnit struct {
	bytes []byte
}

type memoryStorage struct {
	data *sync.Map
	srv  http.Server
}

func OpenBucket(_ context.Context, _ string) (*memoryStorage, error) {
	if len(os.Getenv(HostName)) == 0 || len(os.Getenv(Port)) == 0 {
		return nil, ErrNoSetEnvVars{}
	}

	i := GetMemInstance()

	if !i.isStart() {
		i.initData()
	}

	m := &memoryStorage{
		data: i.getData(),
		srv: http.Server{
			Addr: ":" + os.Getenv(Port),
		},
	}

	mux := http.NewServeMux()
	mux.Handle(pattern, http.StripPrefix(pattern, m))

	if !i.isStart() {
		go func() {
			err := http.ListenAndServe(m.srv.Addr, mux)
			if err != nil {
				log.Println("http Listen err", err)
				panic("http Listen panic")
			}
		}()
		i.setStart()
	}

	return m, nil
}

func (m *memoryStorage) Delete(_ context.Context, objName string) error {
	m.data.Delete(objName)

	return nil
}

func (m *memoryStorage) UploadBytes(_ context.Context, fileAsBytes []byte, objName string) error {
	m.data.Store(objName, dataUnit{
		bytes: fileAsBytes,
	})

	return nil
}

func (m *memoryStorage) UploadByChunks(_ context.Context, fileAsRead io.Reader, objName string) error {
	buf := make([]byte, chunkSize)

	wc := bytes.NewBuffer(make([]byte, 0))

	if _, err := io.CopyBuffer(wc, fileAsRead, buf); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	m.data.Store(objName, dataUnit{
		bytes: wc.Bytes(),
	})

	return nil
}

func (m *memoryStorage) DownloadBytes(_ context.Context, objName string) ([]byte, error) {
	data, ok := m.data.Load(objName)

	if !ok {
		return nil, ErrNoSuchObject{}
	}

	dataUnit, ok := data.(dataUnit)

	if !ok {
		return nil, ErrTypeAssertion{}
	}

	return dataUnit.bytes, nil
}

func (m *memoryStorage) DownloadByChunks(_ context.Context, objName string) (io.ReadCloser, error) {
	data, ok := m.data.Load(objName)

	if !ok {
		return nil, ErrNoSuchObject{}
	}

	dataUnit, ok := data.(dataUnit)

	if !ok {
		return nil, ErrTypeAssertion{}
	}

	return io.NopCloser(bytes.NewReader(dataUnit.bytes)), nil
}

func (m *memoryStorage) GenerateGetObjectSignedURL(_ context.Context, objName string, _ time.Time) (string, error) {
	_, ok := m.data.Load(objName)

	if !ok {
		return "", ErrNoSuchObject{}
	}

	return fmt.Sprintf("http://%v%v%v?%v=%v", os.Getenv(HostName), m.srv.Addr, pattern, urlValue, objName), nil
}
