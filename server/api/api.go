// Package api implements a framework for creating a JSON API.
package api

import (
	"bytes"
	"encoding/json"
	"github.com/boreq/flightradar-backend/logging"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

var log = logging.GetLogger("api")

var InternalServerError = NewError(500, "Internal server error.")
var BadRequest = NewError(400, "Bad request.")

type Error interface {
	GetCode() int
	Error() string
}

func NewError(code int, message string) Error {
	return apiError{code, message}
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err apiError) GetCode() int {
	return err.Code
}

func (err apiError) Error() string {
	return err.Message
}

type Handle func(r *http.Request, p httprouter.Params) (interface{}, Error)

func Call(w http.ResponseWriter, r *http.Request, p httprouter.Params, handle Handle) error {
	code := 200
	response, apiErr := handle(r, p)
	if apiErr != nil {
		response = apiError{apiErr.GetCode(), apiErr.Error()}
		code = apiErr.GetCode()
	}
	j, err := json.Marshal(response)
	if err != nil {
		log.Printf("Marshal error: %s", err)
		j, _ = json.Marshal(InternalServerError)
		code = InternalServerError.GetCode()
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_, err = bytes.NewBuffer(j).WriteTo(w)
	return err
}

func Wrap(handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		Call(w, r, p, handle)
	}
}
