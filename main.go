package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
)

var (
	projectID  string
	apiCred    string
	bucketName string
	folderName string
)

func init() {
	errs := make([]error, 0)
	if apiCred = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); apiCred == "" {
		errs = append(errs, errors.New("GOOGLE_APPLICATION_CREDENTIALS is not configured"))
	}
	if projectID = os.Getenv("PROJECT_ID"); projectID == "" {
		errs = append(errs, errors.New("PROJECT_ID is not configured"))
	}
	if bucketName = os.Getenv("BUCKET_NAME"); bucketName == "" {
		errs = append(errs, errors.New("BUCKET_NAME is not configured"))
	}
	if folderName = os.Getenv("FOLDER_NAME"); folderName == "" {
		errs = append(errs, errors.New("FOLDER_NAME is not configured"))
	}
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		os.Exit(1)
	}
}

func main() {
	filename := ""
	flag.StringVar(&filename, "file", "", "filename to upload")
	flag.Parse()
	if filename == "" {
		exitErrorf("filename must be specified")
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_RDONLY, 0666)
	if err != nil {
		exitErrorf("could not open a file: %v", err)
	}
	defer file.Close()
	_, path := filepath.Split(filename)
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		exitErrorf("failed to create client: %v", err)
	}
	bucket := client.Bucket(bucketName)
	object := filepath.Join(folderName, path)
	wc := bucket.Object(object).NewWriter(ctx)
	if _, err = io.Copy(wc, file); err != nil {
		exitErrorf("could not copy file content: %v", err)
	}
	if err = wc.Close(); err != nil {
		exitErrorf("could not close storage writer: %v", err)
	}
	fmt.Printf("successfully uploaded %q to %q at %v\n", filename, bucketName, wc.Attrs().Updated.String())
	fmt.Printf("download url %s\n", wc.Attrs().MediaLink)
	fmt.Printf("preview url http://storage.googleapis.com/%s/%s\n", bucketName, object)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
