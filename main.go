package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mikerybka/util"
)

func main() {
	port := util.EnvVar("PORT", "2069")
	addr := ":" + port
	err := http.ListenAndServe(addr, &SchemaCafe{"data"})
	if err != nil {
		fmt.Println(err)
		return
	}
}

type SchemaCafe struct {
	DataDir string
}

func (cafe *SchemaCafe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(cafe.DataDir, r.URL.Path)
	if r.Method == "PUT" {
		err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		f, err := os.Create(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		_, err = io.Copy(f, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	if r.Method == "DELETE" {
		err := os.Remove(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}
	fi, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if fi.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, e := range entries {
			fmt.Fprintf(w, "<a href=\"%s\">%s</a>", filepath.Join(r.URL.Path, e.Name()), e.Name())
		}
		return
	}
	f, err := os.Open(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.Copy(w, f)
}
