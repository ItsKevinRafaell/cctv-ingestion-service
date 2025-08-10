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
	uploader  *uploader.LocalUploader
	publisher *mq.RabbitMQPublisher
}

func NewService(uploader *uploader.LocalUploader, publisher *mq.RabbitMQPublisher) Service {
	return &service{uploader: uploader, publisher: publisher}
}

func (s *service) ProcessVideo(file multipart.File, handler *multipart.FileHeader) error {
	// 1. Simpan file
	filePath, err := s.uploader.Save(file, handler)
	if err != nil {
		return err
	}

	// 2. Buat pesan tugas
	taskMessage := map[string]string{"video_path": filePath}

	// 3. Kirim pesan ke antrian
	err = s.publisher.Publish("video_analysis_tasks", taskMessage)
	if err != nil {
		// Di dunia nyata, kita mungkin ingin ada logika retry atau penanganan error lain di sini
		return err
	}

	return nil
}
