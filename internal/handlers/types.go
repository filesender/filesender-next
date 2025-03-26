package handlers

import (
	"time"

	"codeberg.org/filesender/filesender-next/internal/models"
)

type createTransferAPIRequest struct {
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
}

type createTransferAPIResponse struct {
	Transfer models.Transfer `json:"transfer"`
}

type uploadTemplate struct {
	MinDate     string
	DefaultDate string
	MaxDate     string
}

type uploadDoneTemplate struct {
	UserID     string
	TransferID string
	FileCount  int
	BytesSize  int
}

type getTransferTemplate struct {
	FileCount int
	ByteSize  int
	Files     []getTransferTemplateFile
}

type getTransferTemplateFile struct {
	FileName string
	ByteSize int
}
