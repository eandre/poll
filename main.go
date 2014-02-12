package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"net/http"
	"strconv"
)

func main() {
	m := martini.New()
	r := martini.NewRouter()

	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("./static", martini.StaticOptions{Prefix: "/static"}))
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
	}))

	r.Get("/api/votes", websocket.Handler(StreamVotes).ServeHTTP)
	r.Get("/", ServeIndex)
	r.Get("/:id", ServePollHTML)

	m.Action(r.Handle)
	m.Run()
}

func ServeIndex(r render.Render) {
	r.HTML(200, "index", nil)
}

func ServePollHTML(w http.ResponseWriter, req *http.Request, r render.Render, params martini.Params) {
	id, err := strconv.ParseUint(params["id"], 10, 32)
	if err != nil {
		http.Redirect(w, req, "/", 307)
		return
	}

	poll, err := GetPoll(id)
	if err != nil {
		http.Redirect(w, req, "/", 307)
		return
	}

	r.HTML(200, "poll", poll)
}
