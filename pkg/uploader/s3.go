package uploader

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Uploader menyimpan file ke layanan object storage yang kompatibel dengan S3.
type S3Uploader struct {
	Client   *s3.Client
	Bucket   string
	Endpoint string
}

// NewS3Uploader membuat instance baru dari S3Uploader.
func NewS3Uploader(endpoint, accessKey, secretKey, bucket string) (*S3Uploader, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"), // Region bisa apa saja untuk MinIO
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint, SigningRegion: "us-east-1"}, nil
			},
		)),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	// Buat client S3. forcePathStyle penting untuk MinIO.
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &S3Uploader{Client: client, Bucket: bucket, Endpoint: endpoint}, nil
}

// Save mengunggah file ke bucket S3.
func (u *S3Uploader) Save(file multipart.File, handler *multipart.FileHeader) (string, error) {
	// Buat nama file unik
	filename := fmt.Sprintf("%d-%s", time.Now().UnixNano(), handler.Filename)

	_, err := u.Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(u.Bucket),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Hasilkan URL publik palsu untuk tujuan logging (di produksi ini akan berbeda)
	fileURL := fmt.Sprintf("%s/%s/%s", u.Endpoint, u.Bucket, filename)
	return fileURL, nil
}
