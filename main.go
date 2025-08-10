package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

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
	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.Mkdir(uploadPath, os.ModePerm)
	}

	// 5. Buat file baru di server.
	// Untuk menghindari nama file yang sama, kita bisa tambahkan timestamp nanti.
	dst, err := os.Create(filepath.Join(uploadPath, handler.Filename))
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

	// TODO: Langkah selanjutnya adalah menaruh "pesan tugas" ke Message Queue di sini.

	// Kirim respons sukses.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("File %s berhasil diunggah.", handler.Filename)))
}

func main() {
	http.HandleFunc("/ingest/video", videoIngestHandler)

	// Jalankan server di port yang berbeda agar tidak bentrok dengan Backend Utama.
	port := "8081"
	fmt.Printf("Server penerima video (Ingestion Service) berjalan di http://localhost:%s\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Gagal memulai server:", err)
	}
}
