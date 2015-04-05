package web

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Clock interface {
	Now() time.Time
}

//go:generate counterfeiter . WordCounter
type WordCounter interface {
	Count(word string, since time.Time) (uint, error)
}

type handler struct {
	wordCounter WordCounter
	keywords    []string
	clock       Clock
}

func New(wordCounter WordCounter, keywords []string, clock Clock) *mux.Router {
	api := &handler{
		wordCounter: wordCounter,
		keywords:    keywords,
		clock:       clock,
	}
	r := mux.NewRouter()
	r.HandleFunc("/wordcount/{period}", api.handleWordCount).
		Methods("GET")
	return r
}

func (h *handler) handleWordCount(w http.ResponseWriter, req *http.Request) {
	wordCounts := make(map[string]uint)
	for _, keyword := range h.keywords {
		count, err := h.wordCounter.Count(keyword, h.clock.Now().AddDate(0, 0, -1))
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		wordCounts[keyword] = count
	}
	wordCountBytes, err := json.Marshal(wordCounts)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header()["Content-Type"] = []string{"application/json"}
	_, err = w.Write(wordCountBytes)
	if err != nil {
		log.Println(err)
	}
}
