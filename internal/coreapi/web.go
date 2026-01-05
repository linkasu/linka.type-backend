package coreapi

import (
	"bytes"
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"time"
)

//go:embed web/* web/assets/* web/admin/*
var webFS embed.FS

var (
	webRootFS   fs.FS
	assetsRootF fs.FS
)

func init() {
	var err error
	webRootFS, err = fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	assetsRootF, err = fs.Sub(webFS, "web/assets")
	if err != nil {
		panic(err)
	}
}

func assetsHandler() http.Handler {
	fileServer := http.FileServer(http.FS(assetsRootF))
	return http.StripPrefix("/assets/", cacheControl("public, max-age=3600", fileServer))
}

func serveWebFile(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(webRootFS, name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(name))
		}
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeStatusOK(w)
}

func cacheControl(value string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", value)
		next.ServeHTTP(w, r)
	})
}
