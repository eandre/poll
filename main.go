package main

import (
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

func templateAdd(a int, b int) int { return a + b }

func main() {
	// Create a dummy poll
	poll, _ := NewPoll("Hello?", "Are you still there?", "I don't hate you")
	AddPoll(poll)

	m := martini.New()
	r := martini.NewRouter()

	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Use(render.Renderer())
	m.Use(martini.Static("./static", martini.StaticOptions{Prefix: "/static"}))

	r.Post("/api/poll", CreatePoll)
	r.Post("/api/poll/:id", VotePoll)
	r.Get("/api/poll/:id", PollInfo)
	r.Get("/api/poll/:id/stream", VoteStream)
	r.Get("/**", ServeIndex)

	m.Action(r.Handle)
	m.Run()
}

func ServeIndex(r render.Render) {
	r.HTML(200, "base", nil)
}

func PollInfo(r render.Render, params martini.Params) {
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		r.JSON(400, err.Error())
		return
	}

	poll, err := GetPoll(id)
	if err != nil {
		log.Println("Could not find poll:", err)
		r.JSON(404, err.Error())
		return
	}

	r.JSON(200, poll)
}

func VotePoll(w http.ResponseWriter, req *http.Request, params martini.Params) {
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		http.Redirect(w, req, "/", 302)
		return
	}

	poll, err := GetPoll(id)
	if err != nil {
		log.Println("Could not find poll", id, err)
		http.Redirect(w, req, "/", 302)
		return
	}

	if err = req.ParseForm(); err != nil {
		log.Println("Could not parse form:", err)
		http.Redirect(w, req, req.URL.Path, 302)
		return
	}

	// Parse answers
	var answers []uint32
	for _, ids := range req.Form["answer"] {
		id, err := strconv.ParseUint(ids, 10, 32)
		if err != nil {
			log.Println("Could not parse answer id:", err)
			http.Redirect(w, req, req.URL.Path, 302)
			return
		}
		answers = append(answers, uint32(id)-1)
	}

	if err = poll.RecordAnswers(answers...); err != nil {
		log.Println("Could not record answers:", answers, err)
		http.Redirect(w, req, req.URL.Path, 302)
		return
	} else {
		log.Println("Successfully voted for answers", answers)
	}

	http.Redirect(w, req, req.URL.Path, 302)
}

func CreatePoll(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Println("Could not parse form:", err)
		http.Redirect(w, req, req.URL.Path, 302)
		return
	}

	poll, err := NewPoll(req.Form.Get("question"), req.Form["answer"]...)
	if err != nil {
		log.Println("Could not create poll:", err)
		http.Redirect(w, req, req.URL.Path, 302)
		return
	} else {
		pollId := AddPoll(poll)
		log.Println("Successfully created poll with id:", pollId)
		http.Redirect(w, req, fmt.Sprintf("/%d", pollId), 302)
		return
	}
}

func VoteStream(w http.ResponseWriter, req *http.Request, params martini.Params) {
	ws, err := websocket.Upgrade(w, req, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}

	defer ws.Close()

	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		log.Println(err)
		return
	}

	StreamVotes(ws, id)
}
