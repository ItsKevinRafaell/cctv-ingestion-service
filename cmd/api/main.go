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
	// Inisialisasi komponen
	localUploader := uploader.NewLocalUploader("./uploads")
	rabbitPublisher, err := mq.NewRabbitMQPublisher("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Gagal terhubung ke RabbitMQ: %v", err)
	}
	defer rabbitPublisher.Close()
	log.Println("✅ Berhasil terhubung ke RabbitMQ!")

	// Dependency Injection
	ingestService := ingest.NewService(localUploader, rabbitPublisher)
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
