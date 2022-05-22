package main

import (
	"context"
	"fmt"
	"git.epam.com/epm-gdsp/cloud-uploader-lab/azure"
	"io"
	"os"
	"strings"
	"time"
)

const azureFileName = "your file name"

func main() {
	fmt.Println("Hello1")

	//you can get it on azure portal
	os.Setenv("ACCOUNT_KEY", "your key")
	os.Setenv("ACCOUNT_NAME", "your account name")

	//open a bucket for working with azure cloud
	//a bucket also called a container
	bucket, err := azure.OpenBucket(context.Background(), "testcontainer")
	if err != nil {
		return
	}

	//upload  "Hello world" to the hello.txt file
	bucket.UploadBytes(context.Background(), []byte("Hello world"), "hello.txt")

	//upload  "Hello world" to the hello.txt file by chunks
	bucket.UploadByChunks(context.Background(), strings.NewReader("Hello world by chunks"), "chunks.txt")

	//delete hello.txt file
	bucket.Delete(context.Background(), "hello.txt")

	//get bytes from hello.txt
	get, err := bucket.DownloadBytes(context.Background(), "hello.txt")
	if err != nil {
		return
	}

	fmt.Println(string(get))

	reader, err := bucket.DownloadByChunks(context.Background(), "hello.txt")
	if err != nil {
		return
	}
	//reading bytes from the io.ReadCloser
	b := make([]byte, 2)
	defer reader.Close()
	for {
		n, err := reader.Read(b)
		if err == io.EOF {
			break
		}
		if n > 0 {
			fmt.Println(string(b))
		}

	}

	// generate SignedURL
	url, err := bucket.GenerateGetObjectSignedURL(context.Background(), azureFileName, time.Now().Add(time.Hour))

	if err != nil {
		return
	}

	fmt.Println(url)

}
