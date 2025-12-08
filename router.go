// router.go
package main

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"

	SherryServer "github.com/asccclass/sherryserver"
)

func NewRouter(srv *SherryServer.Server, documentRoot string, templateRoot string) *http.ServeMux {
	router := http.NewServeMux()

	// Static File server
	// staticfileserver := SherryServer.StaticFileServer{documentRoot, "index.html"}
	// staticfileserver.AddRouter(router)

	// Custom Index Handler
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleIndex(w, r, documentRoot, templateRoot)
	})

	// App router
	router.HandleFunc("POST /api/tts", SpeakFromWeb) // handleTTS)
	router.HandleFunc("POST /speak", SpeakFromWeb)
	router.HandleFunc("GET /api/weather/{location}", handleWeather)

	// Proxy for MessageHub Socket.io
	u, _ := url.Parse("http://localhost:9090")
	proxy := httputil.NewSingleHostReverseProxy(u)
	router.Handle("/ws/", proxy)

	/*
	   // App router
	   router.HandleFunc("GET /api/notes", GetAll)
	   router.HandleFunc("POST /api/notes", Post)

	   router.Handle("/homepage", oauth.Protect(http.HandlerFunc(Home)))
	   router.Handle("/upload", oauth.Protect(http.HandlerFunc(Upload)))
	*/
	return router
}

func handleIndex(w http.ResponseWriter, r *http.Request, documentRoot string, templateRoot string) {
	// If path is not root or index.html, serve file
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.FileServer(http.Dir(documentRoot)).ServeHTTP(w, r)
		return
	}

	// Check Env Vars
	msgAPI := os.Getenv("MESSAGE_API_URL")
	ttsAPI := os.Getenv("TTS_API_URL")

	data := IndexTemplateData{
		ShowMessageAPI: msgAPI != "",
		ShowTTSAPI:     ttsAPI != "",
		MessageAPIURL:  msgAPI,
	}

	tmplPath := filepath.Join(templateRoot, "index.tpl")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Could not load template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execution failed: "+err.Error(), http.StatusInternalServerError)
	}
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	location := r.PathValue("location")
	if location == "" {
		http.Error(w, "Location is required", http.StatusBadRequest)
		return
	}

	targetURL := "https://www.justdrink.com.tw/apigateway/status/" + location
	resp, err := http.Get(targetURL)
	if err != nil {
		http.Error(w, "Failed to fetch weather data: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	// Copy body
	io.Copy(w, resp.Body)
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
