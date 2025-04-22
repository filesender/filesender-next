// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// DownloadAPI handles `GET /api/v1/download/{userID}/{fileID}`
func DownloadAPI(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		if !file.Chunked {
			file.DownloadCount++
			err = file.Save(stateDir)
			if err != nil {
				slog.Error("Failed increasing download count on file", "error", err, "userID", userID, "fileID", fileID)
				sendError(w, http.StatusInternalServerError, "Failed setting new file meta data")
				return
			}
		}

		filePath := filepath.Join(stateDir, file.UserID, file.FileName)
		sendFile(w, filePath, file.FileName)
	}
}

// ChunkedDownloadAPI handles `GET /api/v1/download/{userID}/{fileID}/{chunk}`
func ChunkedDownloadAPI(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID, chunkStr := r.PathValue("userID"), r.PathValue("fileID"), r.PathValue("chunk")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "user_id", userID, "file_id", fileID, "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "user_id", userID, "file_id", fileID, "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		if !file.Chunked {
			sendError(w, http.StatusNotAcceptable, "This file is not chunked")
			return
		}

		chunk, err := strconv.ParseInt(chunkStr, 10, 0)
		if err != nil {
			slog.Error("Failed converting chunk index to number", "index", chunkStr, "error", err)
			sendError(w, http.StatusBadRequest, "Invalid chunk index")
			return
		}

		if chunk < 0 || int(chunk) >= len(file.Chunks) {
			slog.Error("Chunk index is out of bounds", "index", chunkStr)
			sendError(w, http.StatusBadRequest, "Chunk index is out of bounds")
			return
		}

		chunkOffset := file.Chunks[chunk]

		if int(chunk) == len(file.Chunks)-1 {
			file.DownloadCount++
			err = file.Save(stateDir)
			if err != nil {
				slog.Error("Failed increasing download count on file", "error", err, "userID", userID, "fileID", fileID)
				sendError(w, http.StatusInternalServerError, "Failed setting new file meta data")
				return
			}
		}

		filePath := filepath.Join(stateDir, file.UserID, file.ID, chunkOffset+".bin")
		sendFile(w, filePath, file.FileName)
	}
}

// DownloadInfo handles `HEAD /api/v1/download/{userID}/{fileID}`
func DownloadInfo(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		w.Header().Add("Chunked", strconv.FormatBool(file.Chunked))
		w.Header().Add("Available", strconv.FormatBool(!file.Partial))
		w.Header().Add("Chunk-Count", strconv.FormatInt(int64(len(file.Chunks)), 10))
		w.Header().Add("File-Name", file.EncryptedFileName)

		sendEmptyResponse(w, http.StatusOK)
	}
}

// GetDownloadTemplate handles GET /download/{userID}/{fileID}
func GetDownloadTemplate(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		data := downloadTemplate{
			ByteSize: file.ByteSize,
			UserID:   file.UserID,
			FileID:   file.ID,
		}

		sendTemplate(w, "download", data)
	}
}
