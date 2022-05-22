package main

import (
	"context"
	"fmt"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/gcp"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// For simplicity
func noErr(err error) {
	if err != nil {
		panic(err)
	}
}

const fileUploadBytes = "fileUploadBytes"
const fileUploadByChunks = "fileUploadByChunks"

func getFromURL(url string) string {
	resp, err := http.Get(url)
	noErr(err)
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	noErr(err)
	return string(bytes)
}

// Before running, set the env vars:
// GOOGLE_APPLICATION_CREDENTIALS
// BUCKET
func main() {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	bucket, err := gcp.OpenBucket(ctx, os.Getenv("BUCKET"))
	noErr(err)

	err = bucket.UploadBytes(ctx, []byte("dlfkjdklj"), fileUploadBytes)
	noErr(err)

	err = bucket.UploadByChunks(ctx, strings.NewReader("abcabcabc324"), fileUploadByChunks)
	noErr(err)

	getURL, err := bucket.GenerateGetObjectSignedURL(ctx, fileUploadBytes, time.Now().Add(time.Hour))
	noErr(err)
	fmt.Println("\nPresigned GET URL:", getURL)

	fmt.Println("\nFrom GET URL: ", getFromURL(getURL))

	// err = bucket.Delete(ctx, fileUploadBytes)
	noErr(err)
}
