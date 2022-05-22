package azure

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

type adapter struct {
	ctx context.Context
}

const (
	bufferSize        = 1024 * 1024 // size of the rotating buffers used when uploading
	maxBuffers        = 4           // number of rotating buffers used when uploading
	objectListMaxSize = 10          // set max number of objects listed in ListObjectMethod
)

type adapterInterface interface {
	Delete(bucketName string, objName string) error
	Upload(fileAsBytes []byte, bucketName string, objName string) error
	UploadChunks(fileAsRead io.Reader, bucketName string, objName string) error
	DownloadBytes(bucketName string, objName string) (io.ReadCloser, error)
	GenerateSignedURL(bucketName string, objName string, ttl time.Time) (string, error)
}

func newAdapter(ctx context.Context) (adapterInterface, error) {
	return &adapter{ctx}, nil
}

func createContainerURL(bucketName string) azblob.ContainerURL {

	accountName, accountKey := accountInfo()
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatalf("reading credential error: %v", err)
		return azblob.ContainerURL{}
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", accountName))
	if err != nil {
		log.Fatalf("failed to create bucket with curren credentials: %v ", err)
		return azblob.ContainerURL{}
	}
	serviceURL := azblob.NewServiceURL(*u, p)
	containerURL := serviceURL.NewContainerURL(bucketName)

	_, err = containerURL.Create(context.Background(), azblob.Metadata{}, "container")
	if err != nil {
		log.Fatalf("container creation failed: %v", err)
		return azblob.ContainerURL{}
	}
	return containerURL
}

func createBlobURL(bucketName, objName string) azblob.BlockBlobURL {

	return createContainerURL(bucketName).NewBlockBlobURL(objName)

}

func (a *adapter) ListObjects(bucketName string) ([]string, error) {

	tmp, err := createContainerURL(bucketName).ListBlobsFlatSegment(context.Background(), azblob.Marker{}, azblob.ListBlobsSegmentOptions{
		Details:    azblob.BlobListingDetails{},
		Prefix:     "",
		MaxResults: objectListMaxSize,
	})

	list := make([]string, len(tmp.Segment.BlobItems))

	for i := range tmp.Segment.BlobItems {
		list[i] = tmp.Segment.BlobItems[i].Name
	}

	return list, fmt.Errorf("failed to create list of objects: %v", err)
}

func (a *adapter) Upload(fileAsBytes []byte, bucketName string, objName string) error {

	blobURL := createBlobURL(bucketName, objName)

	_, err := blobURL.Upload(a.ctx, bytes.NewReader(fileAsBytes), azblob.BlobHTTPHeaders{ContentType: http.DetectContentType(fileAsBytes)}, azblob.Metadata{}, azblob.BlobAccessConditions{}, azblob.DefaultAccessTier, nil, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		fmt.Errorf("uploading file error: %w", err)
	}

	return err
}

func (a *adapter) UploadChunks(fileAsRead io.Reader, bucketName string, objName string) error {

	blobURL := createBlobURL(bucketName, objName)

	// Perform UploadStreamToBlockBlob
	bufferSize := bufferSize
	maxBuffers := maxBuffers
	_, err := azblob.UploadStreamToBlockBlob(a.ctx, fileAsRead, blobURL,
		azblob.UploadStreamToBlockBlobOptions{BufferSize: bufferSize, MaxBuffers: maxBuffers})

	return fmt.Errorf("uploading by chunks error: %w", err)
}

func (a *adapter) Delete(bucketName string, objName string) error {

	blobURL := createBlobURL(bucketName, objName)
	_, err := blobURL.Delete(a.ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	return fmt.Errorf("deleting the file error: %w", err)
}

func (a *adapter) DownloadBytes(bucketName string, objName string) (io.ReadCloser, error) {

	blobURL := createBlobURL(bucketName, objName)

	get, err := blobURL.Download(a.ctx, 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, fmt.Errorf("downloading file error: %w", err)
	}

	// Wrap the response body in a ResponseBodyProgress and pass a callback function for progress reporting.
	responseBody := pipeline.NewResponseBodyProgress(get.Body(azblob.RetryReaderOptions{}),
		func(bytesTransferred int64) {
			fmt.Printf("Read %d of %d bytes.", bytesTransferred, get.ContentLength())
		})
	return responseBody, nil
}

func (a *adapter) GenerateSignedURL(bucketName string, objName string, ttl time.Time) (string, error) {

	accountName, accountKey := accountInfo()
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return "", fmt.Errorf("reading credential error: %w", err)
	}

	sasQueryParams, err := azblob.BlobSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,
		ExpiryTime:    ttl,
		ContainerName: bucketName,
		BlobName:      objName,

		Permissions: azblob.BlobSASPermissions{Add: true, Read: true, Write: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		return "", fmt.Errorf("creating query parametrs error: %w", err)
	}

	qp := sasQueryParams.Encode()

	signedURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s?%s",
		accountName, bucketName, objName, qp)

	return signedURL, nil
}
