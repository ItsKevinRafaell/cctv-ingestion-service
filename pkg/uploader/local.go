package uploader

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// LocalUploader menyimpan file ke disk lokal.
type LocalUploader struct {
	UploadPath string
}

func NewLocalUploader(uploadPath string) *LocalUploader {
	return &LocalUploader{UploadPath: uploadPath}
}

func (u *LocalUploader) Save(file multipart.File, handler *multipart.FileHeader) (string, error) {
	if _, err := os.Stat(u.UploadPath); os.IsNotExist(err) {
		os.MkdirAll(u.UploadPath, os.ModePerm)
	}

	// Buat nama file unik untuk menghindari tabrakan
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)
	filePath := filepath.Join(u.UploadPath, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filePath, nil
}
