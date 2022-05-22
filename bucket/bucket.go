package bucket

import (
	"context"
	"io"
	"time"
)

type Bucket interface {
	Delete(ctx context.Context, objName string) error
	UploadBytes(ctx context.Context, fileAsBytes []byte, objName string) error
	UploadByChunks(ctx context.Context, fileAsRead io.Reader, objName string) error
	DownloadBytes(ctx context.Context, objName string) ([]byte, error)
	DownloadByChunks(ctx context.Context, objName string) (io.ReadCloser, error)
	GenerateGetObjectSignedURL(ctx context.Context, objName string, ttl time.Time) (string, error)
}

/*
 Notes:
 1. Each package(AWS,GCP,Azure) must contain function OpenBucket with the following signature:
	OpenBucket(SOME_ARGUMENTS) (Bucket_implementation, error)
 This function returns a structure that satisfies the Bucket interface.

 2. The cloud provider is determined by the package used, which has the OpenBucket function.
		Example:

		var b Bucket = GCP.OpenBucket(ctx,bucketName)

 3. The bucket name is defined in the OpenBucket function.

 4. Please add this line to your implementation

	var _ bucket.Bucket = (*Bucket_implementation)(nil)

 Bucket_implementation is type struct that satisfies the Bucket interface. For instance bucketGCP instead of
 Bucket_implementation.

*/
