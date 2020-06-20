package main

import (
	"github.com/labstack/echo"
	"net/http"
)

var DefaultBadRequestResponse = echo.NewHTTPError(http.StatusBadRequest, map[string]string{
	"error": "Bad Request: check parameters",
})

var DefaultServerErrorResponse = echo.NewHTTPError(http.StatusInternalServerError, map[string]string{
	"error": "Fatal Error occurred at internal service",
})

type MergedResponse struct {
	People []Person `json:"people"`
	Images []Image  `json:"images"`
	Faces  []Face   `json:"faces"`
}

type UploadInitiateResponse struct {
	Bucket        string `json:"bucket"`
	Authorization string `json:"authorization"`
	Policy        string `json:"policy"`
	Filename      string `json:"filename"`
}
