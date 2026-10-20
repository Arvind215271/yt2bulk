package export

import (
	"net/http"
	"github.com/Arvind215271/yt2bulk/internal/api"
)

var (
	ErrInvalidFormat    = api.NewError("invalid_format", "Unsupported export format. Supported formats: JSON", http.StatusBadRequest)
	ErrInvalidInputType = api.NewError("invalid_input_type", "Invalid YouTube input type. Ensure it is a valid video/playlist URL", http.StatusBadRequest)
	ErrInputRequired    = api.NewError("input_required", "Input URL(s) are required to perform the export", http.StatusBadRequest)
)
