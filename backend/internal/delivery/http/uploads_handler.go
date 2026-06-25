package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

// NewUploadsHandler registers routes for file uploads
func NewUploadsHandler(r *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	g := r.Group("/uploads")
	g.Use(authMiddleware)
	{
		g.POST("/evidence", uploadFile("evidence"))
		g.POST("/speedtest", uploadFile("speedtest"))
	}
}

func uploadFile(subDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		ext := filepath.Ext(file.Filename)
		if ext == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File must have a valid extension"})
			return
		}

		uniqueFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		dir := "./uploads/" + subDir
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		filePath := filepath.Join(dir, uniqueFilename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
			return
		}

		fileInfo, err := os.Stat(filePath)
		var size int64
		if err == nil {
			size = fileInfo.Size()
		}

		contentType := file.Header.Get("Content-Type")
		fileURL := fmt.Sprintf("/static/uploads/%s/%s", subDir, uniqueFilename)

		c.JSON(http.StatusOK, gin.H{
			"file_url":     fileURL,
			"filename":     file.Filename,
			"content_type": contentType,
			"size":         size,
		})
	}
}
