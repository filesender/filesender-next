package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/hash"
	"codeberg.org/filesender/filesender-next/internal/id"
)

// UploadAPI handles `POST /upload` & `PATCH /upload/{fileID}`
func UploadAPI(appRoot string, authModule auth.Auth, stateDir string, maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authModule.UserAuth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}

		userID, err = hash.ToBase64(userID)
		if err != nil {
			slog.Info("failed hashing user ID", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed creating user ID")
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			sendError(w, http.StatusRequestEntityTooLarge, "Upload file size too large")
			return
		}

		continued := true
		var uploadOffset int64
		fileID := r.PathValue("fileID")
		if fileID != "" { // upload continues from already existing
			offsetStr := r.Header.Get("Upload-Offset")
			if offsetStr == "" {
				slog.Info("Missing upload offset")
				sendError(w, http.StatusBadRequest, "Missing offset")
				return
			}

			uploadOffset, err = strconv.ParseInt(offsetStr, 10, 64)
			if err != nil || uploadOffset == 0 {
				slog.Info("Invalid upload offset", "offset", offsetStr)
				sendError(w, http.StatusBadRequest, "Invalid offset")
				return
			}
		} else { // new upload
			continued = false
			fileID, err = id.New()
			if err != nil {
				slog.Error("Failed creating file ID", "error", err)
				sendError(w, http.StatusInternalServerError, "Failed to create a random file ID!")
				return
			}
		}

		complete := true
		if completes := r.Header.Get("Upload-Complete"); completes == "0" {
			complete = false
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			slog.Error("Failed opening file", "error", err)

			if err == http.ErrMissingFile {
				sendError(w, http.StatusBadRequest, "No file")
			} else {
				sendError(w, http.StatusInternalServerError, "Lost the file")
			}
			return
		}
		defer func() {
			if err := file.Close(); err != nil {
				slog.Error("Failed closing file", "error", err)
			}
		}()

		totalFileSize := fileHeader.Size

		if continued {
			totalFileSize, err = PartialFileUpload(stateDir, userID, fileID, file, uploadOffset)
			if err != nil {
				slog.Error("Failed handling file upload", "error", err)
				sendError(w, http.StatusInternalServerError, "Failed handling new file upload")
				return
			}
		} else {
			err = FileUpload(stateDir, userID, fileID, file)
			if err != nil {
				slog.Error("Failed handling file upload", "error", err)
				sendError(w, http.StatusInternalServerError, "Failed handling new file upload")
				return
			}
		}

		if complete {
			err = sendRedirect(w, http.StatusSeeOther, appRoot+"view/"+userID+"/"+fileID, "")
			if err != nil {
				sendError(w, http.StatusInternalServerError, "Failed sending redirect")
			}
		} else {
			sendIncompleteResponse(w, appRoot, fileID, maxUploadSize, totalFileSize)
		}
	}
}
