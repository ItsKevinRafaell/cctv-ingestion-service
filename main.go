package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Global variabel untuk koneksi RabbitMQ
var rabbitConn *amqp.Connection

// videoIngestHandler adalah fungsi yang akan menangani semua unggahan video.
func videoIngestHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Pastikan request menggunakan metode POST.
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	// 2. Batasi ukuran file yang diunggah (misalnya, 50MB) untuk keamanan.
	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)
	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
		http.Error(w, "File terlalu besar", http.StatusBadRequest)
		return
	}

	// 3. Ambil file dari form request. "video_clip" adalah nama field
	//    yang nanti akan dikirim oleh skrip Python.
	file, handler, err := r.FormFile("video_clip")
	if err != nil {
		http.Error(w, "Gagal membaca file dari request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("✅ Menerima file: %s, Ukuran: %d bytes\n", handler.Filename, handler.Size)

	// 4. Siapkan tempat untuk menyimpan file.
	// Kita akan membuat folder "uploads" jika belum ada.
	defer file.Close()
	log.Printf("✅ Menerima file: %s", handler.Filename)
	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.Mkdir(uploadPath, os.ModePerm)
	}
	filePath := filepath.Join(uploadPath, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Gagal membuat file di server", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 6. Salin isi file yang diunggah ke file baru di server.
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Gagal menyimpan file", http.StatusInternalServerError)
		return
	}

	log.Printf("   > File berhasil disimpan di: %s\n", dst.Name())

	// ----- INI BAGIAN BARUNYA: Kirim pesan ke RabbitMQ -----
	err = publishToQueue(filePath)
	if err != nil {
		log.Printf("❌ Gagal mengirim pesan ke RabbitMQ: %v", err)
		// Di aplikasi produksi, kita mungkin ingin ada mekanisme retry di sini
		http.Error(w, "Gagal memproses file", http.StatusInternalServerError)
		return
	}
	log.Println("   > Pesan tugas berhasil dikirim ke antrian.")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("File %s berhasil diunggah dan dijadwalkan untuk analisis.", handler.Filename)))
}

// publishToQueue adalah fungsi untuk mengirim pesan ke RabbitMQ.
func publishToQueue(filePath string) error {
	ch, err := rabbitConn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Deklarasikan nama antrian. Jika belum ada, akan dibuat.
	q, err := ch.QueueDeclare(
		"video_analysis_tasks", // Nama antrian
		true,                   // Durable: antrian akan tetap ada jika RabbitMQ restart
		false,                  // Delete when unused
		false,                  // Exclusive
		false,                  // No-wait
		nil,                    // Arguments
	)
	if err != nil {
		return err
	}

	// Buat pesan yang akan dikirim (dalam format JSON)
	taskMessage := map[string]string{"video_path": filePath}
	body, err := json.Marshal(taskMessage)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Kirim pesannya
	return ch.PublishWithContext(ctx,
		"",     // Exchange
		q.Name, // Routing key (nama antrian)
		false,  // Mandatory
		false,  // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func main() {
	var err error
	rabbitConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Gagal terhubung ke RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()
	log.Println("✅ Berhasil terhubung ke RabbitMQ!")

	http.HandleFunc("/ingest/video", videoIngestHandler)

	port := "8081"

	fmt.Printf("Server penerima video (Ingestion Service) berjalan di http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Gagal memulai server:", err)
	}
}
