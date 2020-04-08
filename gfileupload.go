package gfileupload

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
)

func FromBytes(b []byte, objectName string, bucketName string, isPublic bool) (url string, err error) {
	return FromFile(bytes.NewReader(b), objectName, bucketName, isPublic)
}

func FromRequest(req *http.Request, bucketName string, isPublic bool) (filename string, url string, err error) {
	req.ParseMultipartForm(32)
	file, handler, err := req.FormFile("file")
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	filename = handler.Filename
	url, err = FromFile(file, filename, bucketName, isPublic)
	if err != nil {
		return "", "", err
	}
	return filename, url, err
}

func FromFile(file io.Reader, objectName string, bucketName string, isPublic bool) (url string, err error) {
	if bucketName == "" {
		return "", errors.New("bucketName not specified")
	}
	ctx := context.Background()
	client, err := storage.NewClient(ctx) // guessing this is grpc.. TODO REFACTOR. this should only be initialzed once
	if err != nil {
		return "", err
	}
	defer client.Close()

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(objectName)
	if err != nil {
		return "", err
	}

	w := obj.NewWriter(ctx)
	_, err = io.Copy(w, file)
	if err != nil {
		return "", err
	}

	if err := w.Close(); err != nil {
		return "", err
	}

	if isPublic {
		//set permissions
		acl := obj.ACL()
		if err := acl.Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
			return "", err
		}
	}
	// attrs, err := obj.Attrs(ctx)
	// if err := w.Close(); err != nil {
	// 	return "", err
	// }
	url = "https://storage.googleapis.com/" + bucketName + "/" + objectName
	return url, nil
}
