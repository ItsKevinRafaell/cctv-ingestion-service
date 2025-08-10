package ingest

import (
	"log"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) VideoIngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metode tidak diizinkan", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)
	if err := r.ParseMultipartForm(50 * 1024 * 1024); err != nil {
		http.Error(w, "File terlalu besar", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("video_clip")
	if err != nil {
		http.Error(w, "Gagal membaca file dari request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("✅ Menerima file: %s, Ukuran: %d bytes\n", handler.Filename, handler.Size)

	err = h.service.ProcessVideo(file, handler)
	if err != nil {
		log.Printf("❌ Gagal memproses video: %v\n", err)
		http.Error(w, "Gagal memproses file", http.StatusInternalServerError)
		return
	}

	log.Println("   > File berhasil diproses dan tugas analisis dikirim.")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File berhasil diunggah dan dijadwalkan untuk analisis."))
}
