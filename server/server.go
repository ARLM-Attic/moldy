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

var wc world.CmdChan

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

type Target struct {
	Op        string `json:"Op,omitempty"`
	Target    []int  `json:"Target,omitempty"`
	Name      string `json:"Name,omitempty"`
	Precision int    `json:"Precision,omitempty"`
}

func wsView(ws *websocket.Conn) {
	wc.Subscribe(func(ev interface{}) error {
		return websocket.JSON.Send(ws, ev)
	})
	if err := websocket.JSON.Send(ws, wc.State()); err == nil {
		var targ Target
		for {
			targ = Target{}
			if err := websocket.JSON.Receive(ws, &targ); err != nil {
				fmt.Println(err)
				break
			}
			if targ.Target == nil {
				wc.ClearTargets(targ.Name)
			} else {
				wc.AddTarget(targ.Name, targ.Precision, targ.Target[0], targ.Target[1])
			}
		}
	} else {
		fmt.Println(err)
	}
}

func main() {
	wc = world.New(width, height, 5000, 1, 1)
	for i := 0; i < 3; i++ {
		wc.NewMold(fmt.Sprintf("test%v", i))
	}

	http.HandleFunc("/js", js)
	http.Handle("/ws/view", websocket.Handler(wsView))
	http.HandleFunc("/", index)
	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
