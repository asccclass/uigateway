// router.go
package main

import(
   // "fmt"
   "net/http"
   "github.com/asccclass/sherryserver"
)

func NewRouter(srv *SherryServer.Server, documentRoot string)(*http.ServeMux) {
   router := http.NewServeMux()

   // Static File server
   staticfileserver := SherryServer.StaticFileServer{documentRoot, "index.html"}
   staticfileserver.AddRouter(router)

/*
   // App router
   router.HandleFunc("GET /api/notes", GetAll)
   router.HandleFunc("POST /api/notes", Post)

   router.Handle("/homepage", oauth.Protect(http.HandlerFunc(Home)))
   router.Handle("/upload", oauth.Protect(http.HandlerFunc(Upload)))
*/	
   return router
}