package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lanxizhu/go-playground/utils"
)

const ChunkSize = 1024 * 1024 // 1MB

var (
	uploadDir    = "uploads"
	chunkDir     = filepath.Join(uploadDir, "chunks")
	completedDir = filepath.Join(uploadDir, "completed")
)

func init() {
	_ = os.MkdirAll(chunkDir, 0755)
	_ = os.MkdirAll(completedDir, 0755)
}

// Upload handles file upload
func Upload(c *gin.Context) {
	file, _ := c.FormFile("file")
	if file == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file is nil"})
		return
	}

	if err := c.SaveUploadedFile(file, filepath.Join(uploadDir, file.Filename)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("'%s' upload failed", file.Filename)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("'%s' uploaded!", file.Filename), "path": fmt.Sprintf("/media/%s", file.Filename)})
	return
}

// Status self-check upload status and return uploaded chunks
func Status(c *gin.Context) {
	// TODO: Implement the logic to check upload status
	fileId := c.GetHeader("X-File-Id")
	if fileId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-File-ID header is required"})
		return
	}

	rdb := utils.GetRedisClient()
	key := "upload_status:" + fileId
	uploadedChunks, err := rdb.SMembers(c, key).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get upload status"})
		return
	}

	var completedChunks []int
	for _, chunkStr := range uploadedChunks {
		if num, err := strconv.Atoi(chunkStr); err == nil {
			completedChunks = append(completedChunks, num)
		}
	}

	c.JSON(http.StatusOK, gin.H{"uploaded": completedChunks})
	return
}

// Chunk upload a file chunk
func Chunk(c *gin.Context) {
	// TODO: Implement the logic to upload a file
	fileId := c.GetHeader("X-File-Id")
	chunkNumber, _ := strconv.Atoi(c.GetHeader("X-Chunk-Number"))
	chunkTotal, _ := strconv.Atoi(c.GetHeader("X-Total-Chunks"))

	if fileId == "" || chunkNumber <= 0 || chunkTotal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid headers"})
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	defer func() {
		if err = file.Close(); err != nil {
			// Handle the error, e.g., log it
			println("Error closing file:", err.Error())
		}
	}()

	chunkFileName := fmt.Sprintf("%s_%d.chunk", fileId, chunkNumber)
	chunkPath := filepath.Join(chunkDir, chunkFileName)
	out, err := os.Create(chunkPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chunk file"})
		return
	}
	defer func() {
		if err = out.Close(); err != nil {
			// Handle the error, e.g., log it
			println("Error closing output file:", err.Error())
		}
	}()

	if _, err = io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save chunk"})
		return
	}

	rdb := utils.GetRedisClient()
	// // Record uploaded chunks in Redis
	key := "upload_status:" + fileId
	rdb.SAdd(c, key, chunkNumber)

	c.JSON(http.StatusOK, gin.H{"message": "Chunk uploaded successfully"})
	return
}

// Complete after all chunks are uploaded, merge them into the final file
func Complete(c *gin.Context) {
	// TODO: Implement the logic to complete the chunk upload
	fileId := c.GetHeader("X-File-Id")
	fileName := c.GetHeader("X-File-Name")
	totalChunks, _ := strconv.Atoi(c.GetHeader("X-Total-Chunks"))

	if fileId == "" || fileName == "" || totalChunks <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "md5, fileName, totalChunks are required"})
		return
	}

	key := fmt.Sprintf("upload_status:%s", fileId)

	rdb := utils.GetRedisClient()
	uploadedCount, _ := rdb.SCard(c, key).Result()

	if int(uploadedCount) != totalChunks {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not all chunks have been uploaded"})
		return
	}

	finalFilePath := filepath.Join(completedDir, fileName)
	finalFile, err := os.Create(finalFilePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create final file"})
		return
	}

	defer func() {
		if err = finalFile.Close(); err != nil {
			return
		}
	}()

	for i := 1; i <= totalChunks; i++ {
		chunkFileName := fmt.Sprintf("%s_%d.chunk", fileId, i)
		chunkPath := filepath.Join(chunkDir, chunkFileName)
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open chunk"})
			return
		}

		if _, err = io.Copy(finalFile, chunkFile); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to merge chunk"})
			_ = chunkFile.Close()
			return
		}

		_ = chunkFile.Close()
	}

	for i := 1; i <= totalChunks; i++ {
		chunkFileName := fmt.Sprintf("%s_%d.chunk", fileId, i)
		chunkPath := filepath.Join(chunkDir, chunkFileName)
		_ = os.Remove(chunkPath)
	}

	rdb.Del(c, key)

	c.JSON(http.StatusOK, gin.H{"message": "File merged successfully", "path": finalFilePath})
	return
}
