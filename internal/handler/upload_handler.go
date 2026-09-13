package handler

import (
	"fmt"
	"path/filepath"
	"time"

	"serve-swimming-be/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "No file uploaded", err.Error())
	}

	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".pdf": true, ".webp": true}
	if !allowedExts[ext] {
		return response.Error(c, fiber.StatusBadRequest, "Invalid file format. Allowed: JPG, PNG, WEBP, PDF", nil)
	}

	newFilename := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	savePath := fmt.Sprintf("./uploads/%s", newFilename)

	if err := c.SaveFile(file, savePath); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "Failed to save file", err.Error())
	}

	fileURL := fmt.Sprintf("/uploads/%s", newFilename)
	return response.Success(c, fiber.StatusOK, "File uploaded successfully", fiber.Map{
		"url":      fileURL,
		"filename": newFilename,
	})
}
