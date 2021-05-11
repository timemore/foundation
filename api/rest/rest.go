package rest

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code string `json:"code:omitempty"`

	// We use the term description because it describes the error
	// to the developer rather than a message for the end user.
	Description string `json:"description,omitempty"`

	Fields []ErrorResponseField `json:"fields,omitempty"`
	DocURL string               `json:"doc_url,omitempty"`
}

type ErrorResponseField struct {
	Field       string `json:"field"`
	Code        string `json:"code,omitempty"`
	Description string `json:"desciption,omitempty"`
	DocURL      string `json:"doc_url,omitempty"`
}

type EmptyRequest struct{}

type EmptyResponse struct{}

type Responder struct {
	w http.ResponseWriter
}

func RespondTo(w http.ResponseWriter) Responder { return Responder{w} }

func (r Responder) encodeToJSON(jsonData interface{}) error {
	return json.NewEncoder(r.w).Encode(jsonData)
}

func (r Responder) Error(errorData ErrorResponse, httpStatusCode int) {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(httpStatusCode)
	err := r.encodeToJSON(errorData)
	if err != nil {
		panic(err)
	}
}

func (r Responder) EmptyError(httpStatusCode int) {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(httpStatusCode)
	_, _ = r.w.Write([]byte("{}"))
}

func (r Responder) Success(successData interface{}) {
	if successData == nil {
		r.w.WriteHeader(http.StatusNoContent)
		return
	}
	r.SuccessWithHTTPStatusCode(successData, http.StatusOK)
}

func (r Responder) SuccessWithHTTPStatusCode(successData interface{}, httpStatusCode int) {
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(httpStatusCode)
	err := json.NewEncoder(r.w).Encode(successData)
	if err != nil {
		panic(err)
	}
}
