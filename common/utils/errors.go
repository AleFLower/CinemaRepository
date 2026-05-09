package utils

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapGRPCErrorToUser translates technical gRPC errors into user-friendly messages
func MapGRPCErrorToUser(err error) string {
	if err == nil {
		return ""
	}

	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.Unavailable:
			return "The service is temporarily unavailable. Please try again later."
		case codes.DeadlineExceeded:
			return "The request took too long. Please check your connection."
		case codes.NotFound:
			return "The requested item was not found."
		case codes.Unimplemented:
			return "Feature not yet available."
		}
		// Return server-provided error message (e.g. "Seat already taken")
		return st.Message()
	}

	return "An unexpected error occurred. Our engineers are working on it."
}
