package errors

type Code string

const (
	ErrProjectionNotFound Code = "PROJECTION_NOT_FOUND"
	ErrSeatNotExists      Code = "SEAT_NOT_EXISTS"
	ErrSeatOccupied       Code = "SEAT_OCCUPIED"
	ErrInvalidStrategy    Code = "INVALID_STRATEGY"
	ErrInternal           Code = "INTERNAL"
)

type AppError struct {
	Code    Code
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code Code, msg string) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
	}
}
