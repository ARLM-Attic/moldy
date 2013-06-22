package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"github.com/zond/moldy/world"
	"net/http"
	"text/template"
)

const (
	width  = 600
	height = 400
)

var htmlTemplates = template.Must(template.New("htmlTemplates").ParseGlob("templates/html/*.html"))
var jsTemplates = template.Must(template.New("jsTemplates").ParseGlob("templates/js/*.js"))

var wc = world.New(width, height, 1000, 5)

func or500(w http.ResponseWriter, err error) {
	if err != nil {
		fmt.Fprintln(w, err)
		w.WriteHeader(500)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	or500(w, htmlTemplates.ExecuteTemplate(w, "index.html", nil))
}

func js(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
	or500(w, jsTemplates.ExecuteTemplate(w, "jquery-1.8.1.min.js", nil))
	or500(w, jsTemplates.ExecuteTemplate(w, "murmurhash3_gc.js", nil))
	or500(w, jsTemplates.ExecuteTemplate(w, "app.js", map[string]interface{}{
		"width":  width,
		"height": height,
	}))
}

func wsView(ws *websocket.Conn) {
	wc.Subscribe(func(ev interface{}) error {
		return websocket.JSON.Send(ws, ev)
	})
	if err := websocket.JSON.Send(ws, wc.State()); err == nil {
		var x interface{}
		for {
			if err := websocket.JSON.Receive(ws, &x); err != nil {
				fmt.Println(err)
				break
			}
		}
	} else {
		fmt.Println(err)
	}
}

func main() {
	http.HandleFunc("/js", js)
	http.Handle("/ws/view", websocket.Handler(wsView))
	http.HandleFunc("/", index)
	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
