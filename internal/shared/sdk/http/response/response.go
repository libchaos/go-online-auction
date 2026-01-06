package response

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"

	"github.com/cristiano-pacheco/go-online-auction/pkg/errs"
)

func Error(w http.ResponseWriter, err error) {
	rError := &errs.Error{}
	ok := errors.As(err, &rError)
	if !ok {
		httpStatus := http.StatusInternalServerError
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(Envelope{
			"error": map[string]string{
				"code":    "internal_server_error",
				"message": "Internal server error",
			},
		})
		return
	}

	if rError.Status == 0 {
		rError.Status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rError.Status)
	_ = json.NewEncoder(w).Encode(rError)
}

func JSON(w http.ResponseWriter, status int, envelope Envelope, headers http.Header) error {
	js, err := json.MarshalIndent(envelope, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(js)

	return nil
}
