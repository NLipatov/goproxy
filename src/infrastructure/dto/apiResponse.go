package dto

type ApiResponse[T any] struct {
	Payload      *T     `json:"payload,omitempty"`
	ErrorCode    int    `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
