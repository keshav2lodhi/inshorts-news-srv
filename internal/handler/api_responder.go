package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"inshorts.com/inshorts-news-srv/internal/base"
)

// Response represents standard layout of api responses
//
// Example:
//
//	res := Response[<yourDataType>]{}
type Response[T any] struct {
	Success          bool   `json:"success"`                    // general flag which indicates successful execution of api
	Status           int    `json:"status"`                     // http status code
	ID               string `json:"_id,omitempty"`              // Optional: an id which shall be passed to the caller
	ErrorMsg         string `json:"error,omitempty"`            // Optional: actual error message for the request
	Code             string `json:"code,omitempty"`             // Optional: unique error code for the error
	Data             *T     `json:"data,omitempty"`             // Optional: response data is set as data
	ErrorDescription string `json:"errorDescription,omitempty"` // Optional: detailed description for the error code
	RefNumber        string `json:"refNumber,omitempty"`        // Optional: reference number used for easy debugging on logs
}

// APIResponder used to handle api requests more efficient
type APIResponder[T any] struct {
	requestCtx *fiber.Ctx
	Response   Response[T]
}

// NewAPIResponder initializes the api responder based on the context, gets the base repo and loads metadata
func NewAPIResponder[T any](ctx *fiber.Ctx) *APIResponder[T] {
	return &APIResponder[T]{requestCtx: ctx}
}

// ParseBody parses the request body
func (r *APIResponder[T]) ParseBody(body interface{}) error {
	err := json.Unmarshal(r.requestCtx.Body(), body)
	if err != nil {
		r.Response.setError(base.ParsingRequestMsg, base.CodeInvalidData, http.StatusBadRequest)
		return err
	}
	return nil
}

// Respond responds based on the current response
func (r *APIResponder[T]) Respond() error {
	return r.requestCtx.Status(r.Response.Status).JSON(r.Response)
}

// RespondWithSuccess responds with a successful response with response data
func (r *APIResponder[T]) RespondWithSuccess(status int, data *T) error {
	r.Response.setSuccess(data, status)
	return r.Respond()
}

// RespondWithError responds with an error based on status, code and error message
func (r *APIResponder[T]) RespondWithError(status int, code string, errorMsg string) error {
	r.Response.setError(errorMsg, code, status)
	return r.Respond()
}

// SetSuccess sets Response to success including data
func (r *Response[T]) setSuccess(data *T, status int) {
	*r = Response[T]{
		Status:  status,
		Success: true,
		Data:    data,
	}
}

// SetError sets Response to error
func (r *Response[T]) setError(errmsg string, code string, status int) {
	*r = Response[T]{
		Status:   status,
		Success:  false,
		ErrorMsg: errmsg,
		Code:     code,
	}
}
