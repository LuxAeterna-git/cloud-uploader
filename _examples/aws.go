package main

import (
	"context"
	"fmt"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/aws"
	"net/http"
	"os"
	"strings"
	"time"
)

// For simplicity

const fileName = "test"

func putToURL(url, content string) {
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(content))
	noErr(err)
	resp, err := http.DefaultClient.Do(req)
	noErr(err)
	defer resp.Body.Close()
}

// Before running, set the env vars:
// AWS_REGION
// AWS_ACCESS_KEY_ID
// AWS_SECRET_ACCESS_KEY
// BUCKET
func main() {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	bucket, err := aws.OpenBucket(ctx, os.Getenv("BUCKET"))
	noErr(err)

	err = bucket.UploadByChunks(ctx, strings.NewReader("abcabcabc"), fileName)
	noErr(err)

	getURL, err := bucket.GenerateGetObjectSignedURL(ctx, fileName, time.Now().Add(time.Hour))
	noErr(err)
	fmt.Println("\nPresigned GET URL:", getURL)

	fmt.Println("\nFrom GET URL: ", getFromURL(getURL))

	putURL, err := bucket.GeneratePutObjectSignedURL(ctx, fileName, time.Now().Add(time.Hour))
	noErr(err)
	fmt.Println("\nPresigned PUT URL:", putURL)

	putToURL(putURL, "content2")

	fmt.Println("\nFrom GET URL after using PUT URL: ", getFromURL(getURL))

	err = bucket.Delete(ctx, fileName)
	noErr(err)
}
