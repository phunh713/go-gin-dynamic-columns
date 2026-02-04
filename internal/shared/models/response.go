package models

// Base GORM model
type GormModel struct {
	ID int64 `json:"id" gorm:"primaryKey;column:id"`
}

// Base response structure
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SingleResponse for single item responses
type SingleResponse[T any] struct {
	BaseResponse
	Data *T `json:"data,omitempty"`
}

// ListResponse for list/collection responses
type ListResponse[T any] struct {
	BaseResponse
	Data       []T         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// ErrorResponse for error responses
type ErrorResponse struct {
	BaseResponse
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// Pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Helper constructors
func NewSingleResponse[T any](data *T, message string) *SingleResponse[T] {
	return &SingleResponse[T]{
		BaseResponse: BaseResponse{Success: true, Message: message},
		Data:         data,
	}
}

func NewListResponse[T any](data []T, pagination *Pagination, message string) ListResponse[T] {
	return ListResponse[T]{
		BaseResponse: BaseResponse{Success: true, Message: message},
		Data:         data,
		Pagination:   pagination,
	}
}

func NewErrorResponse(error string, errorMessage string) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{Success: false},
		Error:        error,
		Details: map[string]string{
			"error": errorMessage,
		},
	}
}
