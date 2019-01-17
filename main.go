package main

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Amesh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	data := map[string]string{
		"response_type": "in_channel",
		"text":          "オープンソース！",
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(data)
	w.Write(buf.Bytes())
}
