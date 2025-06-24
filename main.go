package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

type Response struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func (cafe *SchemaCafe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(cafe.DataDir, r.URL.Path)
	switch r.Method {
	case http.MethodPut:
		s := &Schema{}
		json.NewDecoder(r.Body).Decode(s)
		err := util.WriteJSONFile(path, s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodDelete:
		err := os.Remove(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case http.MethodGet:
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
			if util.Accept(r, "text/html") {
				for _, e := range entries {
					fmt.Fprintf(w, "<a href=\"%s\">%s</a>", filepath.Join(r.URL.Path, e.Name()), e.Name())
				}
				return
			}
			data := []DirEntry{}
			for _, e := range entries {
				entry := DirEntry{
					Name: e.Name(),
				}
				if e.IsDir() {
					entry.Type = "dir"
				} else {
					entry.Type = "schema"
				}
				data = append(data, entry)
			}
			json.NewEncoder(w).Encode(Response{
				Type: "dir",
				Data: data,
			})
		} else {
			s := &Schema{}
			f, err := os.Open(path)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewDecoder(f).Decode(s)
			json.NewEncoder(w).Encode(Response{
				Type: "schema",
				Data: s,
			})
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type Schema struct {
	Fields []Field `json:"fields"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
