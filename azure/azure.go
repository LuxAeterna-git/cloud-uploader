package azure

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type bucketAzure struct {
	bucketName string
	newAdapter func(ctx context.Context) (adapterInterface, error)
}

func OpenBucket(ctx context.Context, bucketName string) (bucketAzure, error) {
	return bucketAzure{bucketName: bucketName, newAdapter: newAdapter}, nil
}

func accountInfo() (string, string) {
	return os.Getenv("ACCOUNT_NAME"), os.Getenv("ACCOUNT_KEY")
}

func (c bucketAzure) UploadBytes(ctx context.Context, fileAsBytes []byte, objName string) error {

	a, err := c.newAdapter(ctx)
	if err != nil {
		return fmt.Errorf("initialization adapter error: %w", err)
	}

	err = a.Upload(fileAsBytes, c.bucketName, objName)
	return err
}

func (c bucketAzure) UploadByChunks(ctx context.Context, fileAsRead io.Reader, objName string) error {

	a, err := c.newAdapter(ctx)
	if err != nil {
		return fmt.Errorf("initialization adapter error: %w", err)
	}
	err = a.UploadChunks(fileAsRead, c.bucketName, objName)
	return err
}

func (c bucketAzure) Delete(ctx context.Context, objName string) error {

	a, err := c.newAdapter(ctx)
	if err != nil {
		return fmt.Errorf("initialization adapter error: %w", err)
	}

	err = a.Delete(c.bucketName, objName)

	return err
}

func (c bucketAzure) GenerateGetObjectSignedURL(ctx context.Context, objName string, ttl time.Time) (string, error) {
	a, err := c.newAdapter(ctx)
	if err != nil {
		return "", fmt.Errorf("initialization adapter error: %w", err)
	}
	signedURL, err := a.GenerateSignedURL(c.bucketName, objName, ttl)
	return signedURL, err
}

func (c bucketAzure) DownloadBytes(ctx context.Context, objName string) ([]byte, error) {

	a, err := c.newAdapter(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization adapter error: %w", err)
	}
	resp, err := a.DownloadBytes(c.bucketName, objName)
	if err != nil {
		return nil, fmt.Errorf("reading file from Azure error: %w", err)
	}

	var downloadedData []byte
	downloadedData, err = ioutil.ReadAll(resp)
	err = resp.Close()
	if err != nil {
		return nil, err
	}
	return downloadedData, err
}

func (c bucketAzure) DownloadByChunks(ctx context.Context, objName string) (io.ReadCloser, error) {
	a, err := c.newAdapter(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialization adapter error: %w", err)
	}
	resp, err := a.DownloadBytes(c.bucketName, objName)
	if err != nil {
		return nil, fmt.Errorf("reading file from Azure error: %w", err)
	}
	return resp, nil
}
