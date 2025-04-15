// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// UploadAPI handles POST /api/v1/upload
// Expects `expiry_date` in form data
func UploadAPI(authModule auth.Auth, stateDir string, maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authModule.UserAuth(r)
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

		expiryDate, err := time.Parse("2006-01-02", expiryDates[0])
		if err != nil {
			slog.Error("Failed parsing date", "error", err, "input", expiryDates[0])
			sendJSON(w, http.StatusBadRequest, false, "Invalid date format, expected YYYY-MM-DD", nil)
			return
		}

		uploadComplete := true
		if completes := r.Header.Get("Upload-Complete"); completes == "?0" {
			uploadComplete = false
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
			ByteSize:   fileHeader.Size,
			ExpiryDate: expiryDate,
			Chunked:    !uploadComplete,
			Partial:    !uploadComplete,
		}
		err = FileUpload(stateDir, fileMeta, file, fileHeader.Filename)
		if err != nil {
			slog.Error("Failed handling file upload", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed handling new file upload", nil)
			return
		}

		if !uploadComplete {
			sendIncompleteResponse(w, userID, fileID, maxUploadSize, fileHeader.Size)
			return
		}

		err = sendRedirect(w, http.StatusSeeOther, "../../download/"+userID+"/"+fileMeta.ID, "") // Redirect to `/download/<user_id>/<file_id>`
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed sending redirect")
		}
	}
}

// ChunkedUploadAPI handles PATCH /api/v1/upload/{userID}/{fileID}
func ChunkedUploadAPI(authModule auth.Auth, stateDir string, maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathUserID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		userID, err := authModule.UserAuth(r)
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
		if pathUserID != userID {
			slog.Info("Authenticated user does not match uploaded user", "authenticated", userID, "in url", pathUserID, "file ID", fileID, "error", err)
			sendJSON(w, http.StatusUnauthorized, false, "You're not authorized", nil)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			sendJSON(w, http.StatusRequestEntityTooLarge, false, "Upload file size too large", nil)
			return
		}

		uploadComplete := true
		if completes := r.Header.Get("Upload-Complete"); completes == "?0" {
			uploadComplete = false
		}

		fileMeta, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Info("Could not get file info", "error", err)
			sendJSON(w, http.StatusNotFound, false, "Could not find file info", nil)
			return
		}

		if !fileMeta.Chunked || !fileMeta.Partial {
			slog.Info("File is not chunked or is already fully uploaded", "chunked", fileMeta.Chunked, "partial", fileMeta.Partial)
			sendJSON(w, http.StatusConflict, false, "File is not chunked or already fully uploaded", nil)
			return
		}

		offsetStr := r.Header.Get("Upload-Offset")
		if offsetStr == "" {
			slog.Info("Missing upload offset")
			sendJSON(w, http.StatusBadRequest, false, "Missing offset", nil)
			return
		}

		uploadOffset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil || uploadOffset == 0 {
			slog.Info("Invalid upload offset", "offset", offsetStr)
			sendJSON(w, http.StatusBadRequest, false, "Invalid offset", nil)
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

		err = PartialFileUpload(stateDir, fileMeta, file, uploadOffset)
		if err != nil {
			slog.Error("Failed handling file upload", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed handling new file upload", nil)
			return
		}

		fileMeta.ByteSize += fileHeader.Size
		if uploadComplete {
			fileMeta.Partial = false
		}
		err = fileMeta.Save(stateDir)
		if err != nil {
			slog.Error("Failed saving meta file contents", "userID", userID, "fileID", fileID, "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed saving new data", nil)

		}

		if uploadComplete {
			err = sendRedirect(w, http.StatusSeeOther, "../../download/"+userID+"/"+fileMeta.ID, "")
			if err != nil {
				sendError(w, http.StatusInternalServerError, "Failed sending redirect")
			}
		} else {
			sendIncompleteResponse(w, userID, fileID, maxUploadSize, fileHeader.Size)
		}
	}
}

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

		file.DownloadCount++
		err = file.Save(stateDir)
		if err != nil {
			slog.Error("Failed increasing download count on file", "error", err, "userID", userID, "fileID", fileID)
			sendError(w, http.StatusInternalServerError, "Failed setting new file meta data")
			return
		}

		sendFile(stateDir, w, &file)
	}
}
