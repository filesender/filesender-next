package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// UploadAPI handles POST /api/upload
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

		uploadComplete := true
		if completes := r.Header.Get("Upload-Complete"); completes == "?0" {
			uploadComplete = false
		}

		var chunkSize int64
		if !uploadComplete {
			chunkSizeStr := r.Header.Get("Chunk-Size")
			chunkSize, err = strconv.ParseInt(chunkSizeStr, 10, 0)

			if err != nil {
				slog.Error("Expected chunk-size header, failed parsing to int", "chunk-size", chunkSizeStr, "error", err)
				sendJSON(w, http.StatusBadRequest, false, "Invalid Chunk-Size", nil)
				return
			}
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
			ByteSize:  fileHeader.Size,
			Chunked:   !uploadComplete,
			Partial:   !uploadComplete,
			ChunkSize: chunkSize,
		}

		fileNames := r.MultipartForm.Value["file-name"]
		if len(fileNames) == 1 {
			fileMeta.EncryptedFileName = fileNames[0]
		}

		err = FileUpload(stateDir, userID, fileID, fileMeta, file)
		if err != nil {
			slog.Error("Failed handling file upload", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed handling new file upload", nil)
			return
		}

		if !uploadComplete {
			sendIncompleteResponse(w, fileID, maxUploadSize, fileHeader.Size)
			return
		}

		err = sendRedirect(w, http.StatusSeeOther, "../../download/"+userID+"/"+fileID, "") // Redirect to `/download/<user_id>/<file_id>`
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed sending redirect")
		}
	}
}

// ChunkedUploadAPI handles PATCH /api/upload/{userID}/{fileID}
func ChunkedUploadAPI(authModule auth.Auth, stateDir string, maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileID := r.PathValue("fileID")

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

		if fileMeta.ChunkSize != fileHeader.Size && !uploadComplete {
			slog.Info("Meta chunk size is different from file size", "expected", fileMeta.ChunkSize, "got", fileHeader.Size)
			sendJSON(w, http.StatusBadRequest, false, "Invalid chunk size", nil)
			return
		}

		err = PartialFileUpload(stateDir, userID, fileID, file, uploadOffset)
		if err != nil {
			slog.Error("Failed handling file upload", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed handling new file upload", nil)
			return
		}

		fileMeta.ByteSize += fileHeader.Size
		if uploadComplete {
			fileMeta.Partial = false
		}
		err = fileMeta.Save(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed saving meta file contents", "userID", userID, "fileID", fileID, "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed saving new data", nil)

		}

		if uploadComplete {
			err = sendRedirect(w, http.StatusSeeOther, "../../../download/"+userID+"/"+fileID, "")
			if err != nil {
				sendError(w, http.StatusInternalServerError, "Failed sending redirect")
			}
		} else {
			sendIncompleteResponse(w, fileID, maxUploadSize, fileMeta.ByteSize)
		}
	}
}

// UploadTemplate handles GET /{$}
func UploadTemplate(authModule auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authModule.UserAuth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		userID, err = crypto.HashToBase64(userID)
		if err != nil {
			slog.Info("failed hashing user ID", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed creating user ID", nil)
			return
		}

		sendTemplate(w, "upload", uploadTemplate{
			UserID: userID,
		})
	}
}
