package gcp

import (
	"bytes"
	"context"
	"fmt"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/bucket"
	"io"
	"io/ioutil"
	"time"
)

const chunkSize = 32

type bucketGCP struct {
	bucketName string
	newAdapter func(ctx context.Context) (adapterInterface, error)
}

var _ bucket.Bucket = (*bucketGCP)(nil)

func OpenBucket(_ context.Context, bucketName string) (*bucketGCP, error) {
	isExist, err := isBucketExist(bucketName)
	if err != nil {
		return nil, err
	}
	if !isExist {
		err = createBucket(bucketName)
		if err != nil {
			return nil, err
		}
	}
	return &bucketGCP{bucketName: bucketName, newAdapter: newAdapter}, nil
}

func (bucket *bucketGCP) Delete(ctx context.Context, objName string) error {
	a, err := bucket.newAdapter(ctx)
	if err != nil {
		return err
	}
	defer a.Close()
	return a.Delete(objName, bucket.bucketName)
}

func (bucket *bucketGCP) UploadBytes(ctx context.Context, fileAsBytes []byte, objName string) error {
	data := bytes.NewReader(fileAsBytes)
	a, err := bucket.newAdapter(ctx)
	if err != nil {
		return err
	}
	defer a.Close()
	wc := a.NewWriter(objName, bucket.bucketName)
	if _, err = io.Copy(wc, data); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}
	return nil
}

func (bucket *bucketGCP) DownloadBytes(ctx context.Context, objName string) ([]byte, error) {
	a, err := bucket.newAdapter(ctx)
	if err != nil {
		return nil, err
	}
	defer a.Close()
	rc, err := a.NewReader(objName, bucket.bucketName)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %w", objName, err)
	}
	defer rc.Close()
	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll: %w", err)
	}
	return data, nil
}

func (bucket *bucketGCP) GenerateGetObjectSignedURL(ctx context.Context, objName string, ttl time.Time) (string, error) {
	a, err := bucket.newAdapter(ctx)
	if err != nil {
		return "", err
	}
	defer a.Close()
	opts, err := a.OptsGen(ttl)
	if err != nil {
		return "", err
	}
	u, err := a.SignedURL(bucket.bucketName, objName, opts)
	if err != nil {
		return "", fmt.Errorf("storage.SignedURL: %w", err)
	}
	return u, nil
}

func (bucket *bucketGCP) DownloadByChunks(ctx context.Context, objName string) (io.ReadCloser, error) {
	a, err := bucket.newAdapter(ctx)
	if err != nil {
		return nil, err
	}
	defer a.Close()
	rc, err := a.NewReader(objName, bucket.bucketName)
	if err != nil {
		return nil, fmt.Errorf("Object(%q).NewReader: %w", objName, err)
	}
	return rc, nil
}

func (bucket *bucketGCP) UploadByChunks(ctx context.Context, fileAsReadCloser io.Reader, objName string) error {
	a, err := bucket.newAdapter(ctx)
	if err != nil {
		return err
	}
	defer a.Close()
	wc := a.NewWriter(objName, bucket.bucketName)
	buf := make([]byte, chunkSize)
	if _, err = io.CopyBuffer(wc, fileAsReadCloser, buf); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}
	return nil
}
