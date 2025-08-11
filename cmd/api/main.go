package main

import (
	"cctv-ingestion-service/internal/ingest"
	"cctv-ingestion-service/pkg/mq"
	"cctv-ingestion-service/pkg/uploader"
	"fmt"
	"log"
	"net/http"
)

func main() {
	s3Uploader, err := uploader.NewS3Uploader(
		"http://host.docker.internal:9000", // Endpoint MinIO dari dalam Docker
		"minioadmin",                       // Access Key dari docker-compose
		"minio-secret-key",                 // Secret Key dari docker-compose
		"video-clips",                      // Nama bucket
	)
	if err != nil {
		log.Fatalf("Gagal menginisialisasi S3 Uploader: %v", err)
	}

	rabbitPublisher, err := mq.NewRabbitMQPublisher("amqp://guest:guest@host.docker.internal:5672/")
	if err != nil {
		log.Fatalf("Gagal terhubung ke RabbitMQ: %v", err)
	}
	defer rabbitPublisher.Close()
	log.Println("✅ Berhasil terhubung ke RabbitMQ!")

	// Dependency Injection dengan uploader baru
	ingestService := ingest.NewService(s3Uploader, rabbitPublisher)
	ingestHandler := ingest.NewHandler(ingestService)

	// Routing
	http.HandleFunc("/ingest/video", ingestHandler.VideoIngestHandler)

	// Jalankan Server
	port := "8081"
	fmt.Printf("Server penerima video (Ingestion Service) berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Gagal memulai server:", err)
	}
}
