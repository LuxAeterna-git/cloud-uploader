package main

import (
	"context"
	"fmt"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/mem"
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

// Before running, set the env vars:
// HOSTNAME_MEM_SRV
// PORT_MEM_SRV

func main() {
	err := os.Setenv(mem.HostName, "localhost")
	noErr(err)
	err = os.Setenv(mem.Port, "8080")
	noErr(err)

	ctx, _ := context.WithTimeout(context.Background(), time.Minute)

	bucket, err := mem.OpenBucket(ctx, "bucket")
	noErr(err)

	err = bucket.UploadBytes(ctx, []byte("dlfkjdklj"), fileUploadBytes)
	noErr(err)

	file, err := bucket.DownloadBytes(ctx, fileUploadBytes)
	noErr(err)
	fmt.Println(string(file))

	err = bucket.Delete(ctx, fileUploadBytes)
	noErr(err)

	err = bucket.UploadByChunks(ctx, strings.NewReader("abababa"), fileUploadByChunks)
	noErr(err)

	file, err = bucket.DownloadBytes(ctx, fileUploadByChunks)
	noErr(err)
	fmt.Println(string(file))

	/*fi, err := os.Open("epam_logo.png")
	if err == nil {
		err = bucket.UploadByChunks(ctx, fi, "epam_logo.png")
		noErr(err)
		link, err := bucket.GenerateGetObjectSignedURL(ctx, "epam_logo.png", time.Now().Add(time.Minute))
		noErr(err)
		fmt.Println(link)
	}

	time.Sleep(time.Minute)*/

	/*bucket, err = mem.OpenBucket(ctx, "bucket")
	noErr(err)

	file, err = bucket.DownloadBytes(ctx, fileUploadByChunks)
	noErr(err)
	fmt.Println(string(file))

	time.Sleep(time.Minute)*/
}
