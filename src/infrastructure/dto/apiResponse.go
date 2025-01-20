package dto

type ApiResponse[T any] struct {
	Payload      *T     `json:"payload,omitempty"`
	ErrorCode    int    `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}
