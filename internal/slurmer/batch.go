package slurmer

import (
	"io"
	"path/filepath"
	"text/template"
)

func WriteBatch(out io.Writer, batch *BatchProperties) error {
	tmpl, err := template.ParseFiles(filepath.Join("templates", "batch.tmpl"))
	if err != nil {
		return err
	}
	return tmpl.Execute(out, batch)
}
