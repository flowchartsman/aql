package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type DiskFS struct {
	files map[string]string
}

func (d *DiskFS) AddFile(loc string) {
	if d.files == nil {
		d.files = map[string]string{}
	}
	fullpath, err := filepath.Abs(loc)
	if err != nil {
		panic(err.Error())
	}
	relpath := filepath.Base(fullpath)
	d.files[relpath] = fullpath
}

func (d *DiskFS) Open(filename string) (fs.File, error) {
	log.Printf("requested: %s", filename)
	return os.OpenFile(d.files[filename], os.O_RDONLY, 0o444)
}

func main() {
	log.SetFlags(0)
	if len(os.Args) != 2 {
		log.Fatal("usage: go run <file.go> <directory with index.html and wasm>")
	}

	wasmfs := new(DiskFS)
	wasmfs.AddFile(filepath.Join(os.Args[1], "wasm_exec.js"))
	wasmfs.AddFile(filepath.Join(os.Args[1], "aqlparse.wasm"))

	fserv := http.FileServer(http.FS(wasmfs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			f, err := os.OpenFile(filepath.Join(os.Args[1], "index.html"), os.O_RDONLY, 0o444)
			if err != nil {
				panic("error opening index.html: " + err.Error())
			}
			http.ServeContent(w, r, "index.html", time.Now(), f)
			return
		}
		fserv.ServeHTTP(w, r)
	})

	port := ":8080"
	log.Println("serving at port 8080")
	log.Fatal(http.ListenAndServe(port, nil))
}
