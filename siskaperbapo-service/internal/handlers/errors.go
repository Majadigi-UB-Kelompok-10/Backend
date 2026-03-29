package handlers

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	ErrInvalidInput      ErrorCode = "INVALID_INPUT"
	ErrInvalidDate       ErrorCode = "INVALID_DATE"
	ErrInvalidPagination ErrorCode = "INVALID_PAGINATION"
	ErrFileTooLarge      ErrorCode = "FILE_TOO_LARGE"
	ErrInvalidMimeType   ErrorCode = "INVALID_MIME_TYPE"

	ErrBahanPokokNotFound ErrorCode = "BAHAN_POKOK_NOT_FOUND"
	ErrAreaNotFound       ErrorCode = "AREA_NOT_FOUND"
	ErrDataNotFound       ErrorCode = "DATA_NOT_FOUND"

	ErrDatabaseError ErrorCode = "DATABASE_ERROR"
	ErrQueryFailed   ErrorCode = "QUERY_FAILED"

	ErrCloudinaryUpload ErrorCode = "CLOUDINARY_UPLOAD_ERROR"
	ErrCloudinaryDelete ErrorCode = "CLOUDINARY_DELETE_ERROR"

	ErrInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrFileOperation  ErrorCode = "FILE_OPERATION_ERROR"
)

type AppError struct {
	Code       ErrorCode
	Message    string
	StatusCode int
	Err        error
	Details    map[string]interface{}
}

func (e *AppError) Error() string {
	return e.Message
}

type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *AppError) ToResponse() ErrorResponse {
	return ErrorResponse{
		Code:    string(e.Code),
		Message: e.Message,
		Details: e.Details,
	}
}

func NewAppError(code ErrorCode, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Details:    make(map[string]interface{}),
	}
}

func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

func ErrInvalidInputValue(fieldName, reason string) *AppError {
	err := NewAppError(
		ErrInvalidInput,
		fmt.Sprintf("Input '%s' tidak valid: %s", fieldName, reason),
		http.StatusBadRequest,
		nil,
	)
	return err.WithDetails("field", fieldName)
}

func ErrInvalidDateFormat(providedDate string) *AppError {
	err := NewAppError(
		ErrInvalidDate,
		"Format tanggal salah, gunakan YYYY-MM-DD",
		http.StatusBadRequest,
		nil,
	)
	return err.WithDetails("provided_date", providedDate)
}

func ErrFileTooLargeError(maxSizeMB int) *AppError {
	err := NewAppError(
		ErrFileTooLarge,
		fmt.Sprintf("Ukuran file terlalu besar (max: %dMB)", maxSizeMB),
		http.StatusRequestEntityTooLarge,
		nil,
	)
	return err.WithDetails("max_size_mb", maxSizeMB)
}

func ErrInvalidMimeTypeError(got string, allowed []string) *AppError {
	err := NewAppError(
		ErrInvalidMimeType,
		fmt.Sprintf("Tipe file tidak didukung. Hanya: %v", allowed),
		http.StatusBadRequest,
		nil,
	)
	return err.WithDetails("received", got).WithDetails("allowed", allowed)
}

func ErrBahanPokokNotFoundError(identifier string) *AppError {
	err := NewAppError(
		ErrBahanPokokNotFound,
		fmt.Sprintf("Komoditas '%s' tidak ditemukan", identifier),
		http.StatusNotFound,
		nil,
	)
	return err.WithDetails("identifier", identifier)
}

func ErrAreaNotFoundError(identifier string) *AppError {
	err := NewAppError(
		ErrAreaNotFound,
		fmt.Sprintf("Area '%s' tidak ditemukan", identifier),
		http.StatusNotFound,
		nil,
	)
	return err.WithDetails("identifier", identifier)
}

func ErrDataNotFoundError(dataType string) *AppError {
	err := NewAppError(
		ErrDataNotFound,
		fmt.Sprintf("Data %s tidak ditemukan", dataType),
		http.StatusNotFound,
		nil,
	)
	return err.WithDetails("data_type", dataType)
}

func ErrDatabaseOperationError(operation string, underlyingErr error) *AppError {
	err := NewAppError(
		ErrDatabaseError,
		"Gagal melakukan operasi database",
		http.StatusInternalServerError,
		underlyingErr,
	)
	return err.WithDetails("operation", operation)
}

func ErrCloudinaryUploadError(underlyingErr error) *AppError {
	err := NewAppError(
		ErrCloudinaryUpload,
		"Gagal upload gambar ke Cloudinary",
		http.StatusBadGateway,
		underlyingErr,
	)
	return err
}

func ErrCloudinaryDeleteError(publicID string, underlyingErr error) *AppError {
	err := NewAppError(
		ErrCloudinaryDelete,
		"Gagal menghapus gambar dari Cloudinary",
		http.StatusBadGateway,
		underlyingErr,
	)
	return err.WithDetails("public_id", publicID)
}

func ErrInternalServerError(operation string, underlyingErr error) *AppError {
	err := NewAppError(
		ErrInternalServer,
		"Terjadi kesalahan server internal",
		http.StatusInternalServerError,
		underlyingErr,
	)
	return err.WithDetails("operation", operation)
}


func HandleAppError(logger *Logger, appErr *AppError) map[string]interface{} {
	ctx := map[string]interface{}{
		"error_code": appErr.Code,
		"status":     appErr.StatusCode,
	}
	if len(appErr.Details) > 0 {
		ctx["details"] = appErr.Details
	}

	if appErr.Err != nil {
		logger.Error(appErr.Message, appErr.Err, ctx)
	} else {
		logger.Warn(appErr.Message, ctx)
	}

	return map[string]interface{}{
		"code":    appErr.Code,
		"message": appErr.Message,
		"details": appErr.Details,
	}
}
