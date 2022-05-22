package gcp

import (
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
)

type adapter struct {
	client *storage.Client
	ctx    context.Context
}

// set ur project id
const projectID = "test-obj-store"

func isBucketExist(bucketName string) (bool, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return false, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()
	var buckets []string
	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return false, err
		}
		buckets = append(buckets, battrs.Name)
	}
	for _, bucket := range buckets {
		if bucket == bucketName {
			return true, nil
		}
	}
	return false, nil
}

func createBucket(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	bucket := client.Bucket(bucketName)
	if err := bucket.Create(ctx, projectID, nil); err != nil {
		return fmt.Errorf("Bucket(%q).Create: %v", bucketName, err)
	}
	return nil
}

func (a *adapter) OptsGen(ttl time.Time) (*storage.SignedURLOptions, error) {
	serviceAccount := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	jsonKey, err := ioutil.ReadFile(serviceAccount)
	if err != nil {
		return &storage.SignedURLOptions{}, fmt.Errorf("ioutil.ReadFile: %v", err)
	}
	conf, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		return &storage.SignedURLOptions{}, fmt.Errorf("google.JWTConfigFromJSON: %v", err)
	}
	opts := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         http.MethodGet,
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Expires:        ttl,
	}
	return opts, nil
}

type adapterInterface interface {
	io.Closer
	Delete(objName, bucketName string) error
	NewWriter(objName, bucketName string) io.WriteCloser
	NewReader(objName, bucketName string) (io.ReadCloser, error)
	SignedURL(bucket string, object string, opts *storage.SignedURLOptions) (string, error)
	OptsGen(ttl time.Time) (*storage.SignedURLOptions, error)
}

func (a *adapter) SignedURL(bucket string, object string, opts *storage.SignedURLOptions) (string, error) {
	return storage.SignedURL(bucket, object, opts)
}

func newAdapter(ctx context.Context) (adapterInterface, error) {
	client, err := storage.NewClient(ctx)
	return &adapter{client, ctx}, err
}

func (a *adapter) Close() error {
	return a.client.Close()
}

func (a *adapter) Delete(objName, bucketName string) error {
	o := a.client.Bucket(bucketName).Object(objName)
	if err := o.Delete(a.ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", objName, err)
	}
	return nil
}

func (a *adapter) NewWriter(objName, bucketName string) io.WriteCloser {
	return a.client.Bucket(bucketName).Object(objName).NewWriter(a.ctx)
}

func (a *adapter) NewReader(objName, bucketName string) (io.ReadCloser, error) {
	return a.client.Bucket(bucketName).Object(objName).NewReader(a.ctx)
}
