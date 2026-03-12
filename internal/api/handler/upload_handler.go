package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const defaultMaxImageMB = 5

var allowedImageMime = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

type UploadHandler struct {
	uploadDir string
	maxBytes  int64
}

func NewUploadHandler(uploadDir string, maxImageMB int) *UploadHandler {
	trimmedDir := strings.TrimSpace(uploadDir)
	if trimmedDir == "" {
		trimmedDir = "./data/uploads"
	}
	if maxImageMB <= 0 {
		maxImageMB = defaultMaxImageMB
	}

	return &UploadHandler{
		uploadDir: trimmedDir,
		maxBytes:  int64(maxImageMB) * 1024 * 1024,
	}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		writeError(c, http.StatusBadRequest, "missing file")
		return
	}
	if fileHeader.Size <= 0 {
		writeError(c, http.StatusBadRequest, "empty file")
		return
	}
	if fileHeader.Size > h.maxBytes {
		writeError(c, http.StatusBadRequest, "file too large")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		writeInternalError(c)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	header := make([]byte, 512)
	n, readErr := io.ReadFull(file, header)
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		writeError(c, http.StatusBadRequest, "invalid file")
		return
	}

	mimeType := http.DetectContentType(header[:n])
	ext, ok := allowedImageMime[mimeType]
	if !ok {
		writeError(c, http.StatusBadRequest, "unsupported image type")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeInternalError(c)
		return
	}

	now := time.Now().UTC()
	relDir := filepath.Join("images", now.Format("2006"), now.Format("01"), now.Format("02"))
	absDir := filepath.Join(h.uploadDir, relDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		writeInternalError(c)
		return
	}

	randomSuffix, err := randomHex(8)
	if err != nil {
		writeInternalError(c)
		return
	}

	filename := fmt.Sprintf("%d_%s%s", now.UnixNano(), randomSuffix, ext)
	absolutePath := filepath.Join(absDir, filename)

	dst, err := os.Create(absolutePath)
	if err != nil {
		writeInternalError(c)
		return
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err := io.Copy(dst, file); err != nil {
		writeInternalError(c)
		return
	}

	url := path.Join("/uploads", filepath.ToSlash(relDir), filename)

	c.JSON(http.StatusCreated, gin.H{
		"url":  url,
		"mime": mimeType,
		"size": fileHeader.Size,
	})
}

func randomHex(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
