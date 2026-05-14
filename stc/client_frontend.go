package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static/*
var stcStaticFiles embed.FS

func (c *STCClientPlugin) serveFrontend(w http.ResponseWriter, r *http.Request) {
	sub, err := fs.Sub(stcStaticFiles, "static")
	if err != nil {
		http.Error(w, "frontend unavailable", http.StatusInternalServerError)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/static/") {
		http.StripPrefix("/static/", http.FileServer(http.FS(sub))).ServeHTTP(w, r)
		return
	}
	http.ServeFileFS(w, r, sub, "index.html")
}
