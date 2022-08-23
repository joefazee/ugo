package ugo

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

// WriteJson write JSON to http.ResponseWriter
func (u *Ugo) WriteJson(w http.ResponseWriter, statusCode int, data interface{}, headers ...http.Header) error {

	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_, err = w.Write(out)

	if err != nil {
		return err
	}
	return nil
}

func (u *Ugo) ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 // 1mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single jason value")
	}

	return nil
}

func (u *Ugo) WriteXML(w http.ResponseWriter, statusCode int, data interface{}, headers ...http.Header) error {

	out, err := xml.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(statusCode)

	_, err = w.Write(out)

	if err != nil {
		return err
	}
	return nil
}

func (u *Ugo) DownloadFile(w http.ResponseWriter, r *http.Request, path, fileName string) error {

	fp := filepath.Join(path, fileName)
	fileToServer := filepath.Clean(fp)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; file=\"%s\"", fileName))
	http.ServeFile(w, r, fileToServer)

	return nil
}

func (u *Ugo) Error404(w http.ResponseWriter) {
	u.ErrorStatus(w, http.StatusNotFound)
}

func (u *Ugo) Error500(w http.ResponseWriter) {
	u.ErrorStatus(w, http.StatusInternalServerError)
}

func (u *Ugo) ErrorUnauthorized(w http.ResponseWriter) {
	u.ErrorStatus(w, http.StatusUnauthorized)
}

func (u *Ugo) ErrorForbidden(w http.ResponseWriter) {
	u.ErrorStatus(w, http.StatusForbidden)
}

func (u *Ugo) ErrorStatus(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
