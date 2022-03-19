package slurmer

import (
	"encoding/json"
	"github.com/ShinoYasx/Slurmer/pkg/slurmer"
	"io"
	"net/http"
	"path/filepath"
	"text/template"
)

func Response(w http.ResponseWriter, v interface{}, code int) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		Error(w, http.StatusInternalServerError)
		panic(err)
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	if err != nil {
		panic(err)
	}
}

func Error(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

func Ok(w http.ResponseWriter, v interface{}) {
	Response(w, v, http.StatusOK)
}

func WriteBatch(out io.Writer, batch *slurmer.BatchProperties) error {
	tmpl, err := template.ParseFiles(filepath.Join("templates", "batch.tmpl"))
	if err != nil {
		return err
	}
	return tmpl.Execute(out, batch)
}
