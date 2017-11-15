package httputil

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ReadBytes(r *http.Request) ([]byte, error) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body.Close()
	return bytes, nil
}

func UnmarshalJsonBody(r *http.Request, v interface{}) error {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return err
	}

	return json.Unmarshal(bodyBytes, &v)
}

func Json(ctx context.Context, w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(data)
}
