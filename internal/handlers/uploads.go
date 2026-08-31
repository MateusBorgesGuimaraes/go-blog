package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadHandler struct {
	BaseURL string
}

func NewUploadHandler(baseURL string) *UploadHandler {
	return &UploadHandler{BaseURL: baseURL}
}

const maxUploadSize = 5 << 20

var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true,
}

// UploadImage responde POST /api/admin/uploads
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "arquivo muito grande (máximo 5MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "nenhum arquivo enviado no campo 'image'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedExtensions[ext] {
		http.Error(w, "formato de arquivo não permitido (use jpg, png, webp ou gif)", http.StatusBadRequest)
		return
	}

	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destPath := filepath.Join("uploads", filename)

	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "erro ao salvar imagem", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "erro ao salvar imagem", http.StatusInternalServerError)
		return
	}

	imageURL := fmt.Sprintf("%s/uploads/%s", h.BaseURL, filename)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(fmt.Appendf(nil, `{"url":"%s"}`, imageURL))
}
