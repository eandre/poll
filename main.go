package main

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func templateAdd(a int, b int) int { return a + b }

func main() {
	runtime.GOMAXPROCS(4)

	go func() {
		for {
			answer := uint32(rand.Intn(len(poll.Answers)))
			poll.RecordAnswers(answer)
			duration := time.Duration(rand.Intn(500)) * time.Millisecond
			time.Sleep(duration)
		}
	}()

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

func VotePoll(req *http.Request, r render.Render, params martini.Params) {
	id, err := strconv.ParseUint(params["id"], 10, 64)
	if err != nil {
		r.JSON(400, err.Error())
		return
	}

	poll, err := GetPoll(id)
	if err != nil {
		log.Println("Could not find poll", id, err)
		r.JSON(404, err.Error())
		return
	}

	if err = req.ParseForm(); err != nil {
		log.Println("Could not parse form:", err)
		r.JSON(400, err.Error())
		return
	}

	// Parse answers
	var answers []uint32
	for _, ids := range req.PostForm["answer[]"] {
		id, err := strconv.ParseUint(ids, 10, 32)
		if err != nil {
			log.Println("Could not parse answer id:", err)
			r.JSON(400, err.Error())
			return
		}
		answers = append(answers, uint32(id)-1)
	}

	if err = poll.RecordAnswers(answers...); err != nil {
		log.Println("Could not record answers:", answers, err)
		r.JSON(400, err.Error())
		return
	} else {
		log.Println("Successfully voted for answers", answers)
	}
	r.JSON(200, nil)
}

func CreatePoll(req *http.Request, r render.Render) {
	if err := req.ParseForm(); err != nil {
		log.Println("Could not parse form:", err)
		r.JSON(400, err.Error())
		return
	}

	multipleChoice := req.PostForm.Get("multipleChoice") == "true"
	poll, err := NewPoll(multipleChoice, req.PostForm.Get("question"),
		req.PostForm["answer[]"]...)
	if err != nil {
		log.Println("Could not create poll:", err)
		r.JSON(400, err.Error())
		return
	} else {
		pollId := AddPoll(poll)
		log.Println("Successfully created poll with id:", pollId)
		r.JSON(200, pollId)
		return
	}
}

func VoteStream(w http.ResponseWriter, req *http.Request, params martini.Params) {
	ws, err := websocket.Upgrade(w, req, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		log.Println("Could not handshake:", err.Error())
		http.Error(w, "Not a websocket handshake", 401)
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
