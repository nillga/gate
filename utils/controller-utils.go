package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/nillga/jwt-server/errors"
)

func InternalServerError(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusInternalServerError, err)
}

func BadRequest(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusBadRequest, err)
}

func Unauthorized(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusUnauthorized, err)
}

func NotFound(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusNotFound, err)
}

func BadGateway(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusBadGateway, err)
}

func Forbidden(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusForbidden, err)
}

func UnprocessableEntity(w http.ResponseWriter, err error) {
	errorSwitch(w, http.StatusUnprocessableEntity, err)
}

func WrongStatus(w http.ResponseWriter, r *http.Response) {
	w.WriteHeader(r.StatusCode)
	if _, err := io.Copy(w, r.Body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed parsing error to JSON: initial error code: ", r.StatusCode)
	}
}

func errorSwitch(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errors.ProceduralError{Message: err.Error()})
}
