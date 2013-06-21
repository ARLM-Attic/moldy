package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
	"text/template"
)

const (
	width  = 1024
	height = 800
)

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))

func or500(w http.ResponseWriter, err error) {
	if err != nil {
		fmt.Fprintln(w, err)
		w.WriteHeader(500)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	or500(w, htmlTemplates.ExecuteTemplate(w, "index.html", nil))
}

func js(w http.ResponseWriter, r *http.Request) {
	or500(w, jsTemplates.ExecuteTemplate(w, "jquery-1.8.1.min.js", nil))
	or500(w, jsTemplates.ExecuteTemplate(w, "app.js", map[string]interface{}{
		"width":  width,
		"height": height,
	}))
}

func ws(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func main() {
	http.HandleFunc("/js", js)
	http.Handle("/ws", websocket.Handler(ws))
	http.HandleFunc("/", index)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
