// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// UploadAPI handles POST /api/v1/upload
// Expects `expiry_date` in form data
func UploadAPI(maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendJSON(w, http.StatusUnauthorized, false, "You're not authenticated", nil)
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		userID, err = crypto.HashToBase64(userID)
		if err != nil {
			slog.Info("failed hashing user ID", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed creating user ID", nil)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			sendJSON(w, http.StatusRequestEntityTooLarge, false, "Upload file size too large", nil)
			return
		}

		expiryDates := r.MultipartForm.Value["expiry-date"]
		if len(expiryDates) != 1 {
			sendJSON(w, http.StatusBadRequest, false, "Expected an expiry date", nil)
			return
		}

		expiryDate, err := time.Parse("2006-02-03", expiryDates[0])
		if err != nil {
			slog.Error("Failed parsing date", "error", err, "input", expiryDates[0])
			sendJSON(w, http.StatusBadRequest, false, "Invalid date format, expected YYYY-MM-DD", nil)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			slog.Error("Failed opening file", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Lost the file", nil)
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				slog.Error("Failed closing file", "error", err)
			}
		}()

		fileID, err := id.New()
		if err != nil {
			slog.Error("Failed creating file ID", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed to create a random file ID!", nil)
			return
		}

		fileMeta := models.File{
			ID:         fileID,
			UserID:     userID,
			ByteSize:   int(fileHeader.Size),
			ExpiryDate: expiryDate,
		}
		err = FileUpload(fileMeta, file)
		if err != nil {
			slog.Error("Failed handling file upload", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed handling new file upload", nil)
			return
		}

		err = sendRedirect(w, http.StatusSeeOther, "../../upload/"+fileMeta.ID, fileMeta.ID) // Redirect to `/upload/<file_id>`
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed sending redirect")
		}
	}
}

// DownloadAPI handles `GET /api/v1/download/{userID}/{fileID}`
func DownloadAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		file.DownloadCount++
		err = file.Save()
		if err != nil {
			slog.Error("Failed increasing download count on file", "error", err, "userID", userID, "fileID", fileID)
			sendError(w, http.StatusInternalServerError, "Failed setting new file meta data")
			return
		}

		sendFile(w, &file)
	}
}
