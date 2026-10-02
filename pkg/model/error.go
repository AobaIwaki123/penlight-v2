package model

// Standard error codes (The 6 essential codes from ADR-0014)
const (
	CodeInvalidParams       = "INVALID_PARAMS"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeInsufficientMembers = "INSUFFICIENT_MEMBERS"
	CodeInternalError       = "INTERNAL_ERROR"
)

// ParamError records an individual invalid request parameter.
type ParamError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// AppError represents an RFC 9457 Problem Details error payload schema.
type AppError struct {
	Status        int          `json:"status"`
	Code          string       `json:"code"`
	Title         string       `json:"title"`
	Detail        string       `json:"detail"`
	InvalidParams []ParamError `json:"invalid_params,omitempty"`
}

func (e AppError) Error() string {
	return e.Detail
}
