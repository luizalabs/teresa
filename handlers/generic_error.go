package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/luizalabs/tapi/models"
)

// Reason is used to keep the reasons to attach with the error
type Reason string

// List of reasons to attach with the error response
const (
	BadRequest          Reason = "BadRequest"
	Unauthorized        Reason = "Unauthorized"
	Forbidden           Reason = "Forbidden"
	NotFound            Reason = "NotFound"
	Conflict            Reason = "Conflict"
	InternalServerError Reason = "InternalServerError"
)

// const (
// 	TooManyTeams Reason = "TooManyTeams"
// 	TeamNotValid Reason = "TeamNotValid"
//
// // UserNotMemberOfAnyTeam Reason = "UserNotMemberOfAnyTeam"
// )

// GenericError is used to help when returning simple and descritive errors
// to the api
type GenericError struct {
	Payload *models.Error `json:"body,omitempty"`
}

// WithMessage add a message to the error
func (e *GenericError) WithMessage(message string) *GenericError {
	e.Payload.Message = &message
	return e
}

// WithReason add a reason to the error
func (e *GenericError) WithReason(reason Reason) *GenericError {
	e.Payload.Reason = string(reason)
	return e
}

// WriteResponse to fits the interface of middleware.Responder
func (e *GenericError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if e.Payload != nil {
		rw.WriteHeader(int(*e.Payload.Code))
		if err := producer.Produce(rw, e.Payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// NewGenericError is a helper to create a GenericError
func NewGenericError(code int32, message ...interface{}) *GenericError {
	msg := fmt.Sprint(message...)
	return &GenericError{
		Payload: &models.Error{
			Code:    &code,
			Message: &msg,
		},
	}
}

// NewBadRequestError returns a Bad Request (http status code 400) to the Api
func NewBadRequestError(message ...interface{}) *GenericError {
	return NewGenericError(400, message...).WithReason(BadRequest)
}

// NewUnauthorizedError returns a Unauthorized Error (http status code 401) to the Api
func NewUnauthorizedError(message ...interface{}) *GenericError {
	return NewGenericError(401, message...).WithReason(Unauthorized)
}

// NewForbiddenError returns a Forbidden Error (http status code 403) to the Api
func NewForbiddenError() *GenericError {
	return NewGenericError(403, "Forbidden").WithReason(Forbidden)
}

// NewNotFoundError returns a Not Found Error (http status code 404) to the Api
func NewNotFoundError(message ...interface{}) *GenericError {
	return NewGenericError(404, message...).WithReason(NotFound)
}

// NewConflictError returns a Conflict Error (http status code 409) to the Api
func NewConflictError(message ...interface{}) *GenericError {
	return NewGenericError(409, message...).WithReason(Conflict)
}

// func NewUnprocessableEntityError(message ...interface{}, reason Reason) *GenericError {
// 	return NewGenericError(422, message).WithReason(reason)
// }

// NewInternalServerError returns a Internal Server Error (http status code 500) to the Api
func NewInternalServerError(message ...interface{}) *GenericError {
	return NewGenericError(500, message...).WithReason(InternalServerError)
}
