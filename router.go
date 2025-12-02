// router.go
package main

import (
	// "fmt"
	"io"
	"net/http"

	SherryServer "github.com/asccclass/sherryserver"
)

func NewRouter(srv *SherryServer.Server, documentRoot string) *http.ServeMux {
	router := http.NewServeMux()

	// Static File server
	staticfileserver := SherryServer.StaticFileServer{documentRoot, "index.html"}
	staticfileserver.AddRouter(router)

	// App router
	router.HandleFunc("POST /api/tts", handleTTS)

	/*
	   // App router
	   router.HandleFunc("GET /api/notes", GetAll)
	   router.HandleFunc("POST /api/notes", Post)

	   router.Handle("/homepage", oauth.Protect(http.HandlerFunc(Home)))
	   router.Handle("/upload", oauth.Protect(http.HandlerFunc(Upload)))
	*/
	return router
}

func handleTTS(w http.ResponseWriter, r *http.Request) {
	// 1. Forward request to Python Service
	resp, err := http.Post("http://localhost:8880/v1/audio/speech", "application/json", r.Body)
	if err != nil {
		http.Error(w, "Failed to call TTS service: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 2. Copy response headers and body
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	// Copy the audio data
	_, _ = io.Copy(w, resp.Body)
}
