package ingest

import (
	"cctv-ingestion-service/pkg/mq"
	"cctv-ingestion-service/pkg/uploader"
	"mime/multipart"
)

type Service interface {
	ProcessVideo(file multipart.File, handler *multipart.FileHeader) error
}

type service struct {
	uploader  *uploader.S3Uploader
	publisher *mq.RabbitMQPublisher
}

func NewService(uploader *uploader.S3Uploader, publisher *mq.RabbitMQPublisher) Service {
	return &service{uploader: uploader, publisher: publisher}
}

func (s *service) ProcessVideo(file multipart.File, handler *multipart.FileHeader) error {
	// 1. Simpan file dan dapatkan URL publik & path internalnya
	fileURL, filePath, err := s.uploader.Save(file, handler) // Fungsi Save sekarang akan mengembalikan 2 nilai
	if err != nil {
		return err
	}

	// 2. Buat pesan tugas yang berisi kedua alamat
	taskMessage := map[string]string{
		"video_url":  fileURL,
		"video_path": filePath, // Path internal untuk diakses worker
	}

	// 3. Kirim pesan ke antrian
	err = s.publisher.Publish("video_analysis_tasks", taskMessage)
	if err != nil {
		return err
	}

	return nil
}
