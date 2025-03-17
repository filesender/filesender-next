package handlers

import (
	"time"

	"codeberg.org/filesender/filesender-next/internal/models"
)

type createTransferAPIRequest struct {
	Subject    string     `json:"subject,omitempty"`
	Message    string     `json:"message,omitempty"`
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
}

type createTransferAPIResponse struct {
	Transfer models.Transfer `json:"transfer"`
}
