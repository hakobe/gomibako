package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/dimfeld/httptreemux"
	"github.com/urfave/negroni"
)

func _templates() map[string]*template.Template {
	templates := make(map[string]*template.Template)
	names := []string{"index", "inspect"}

	for _, name := range names {
		templates[name] =
			template.Must(template.ParseFiles(
				"templates/layout.tmpl.html", "templates/"+name+".tmpl.html"))
	}

	return templates
}

type gomibakoKey string

type gomibakoAddingReq struct {
	key     string
	request *http.Request
}

type gomibakoRepository struct {
	gomibakos map[gomibakoKey][]*http.Request
	mutex     sync.RWMutex
}

func newGomibakoRepository() *gomibakoRepository {
	gr := gomibakoRepository{
		gomibakos: make(map[gomibakoKey][]*http.Request),
		mutex:     sync.RWMutex{},
	}
	return &gr
}

func (gr *gomibakoRepository) Add(key gomibakoKey, req *http.Request) {
	gr.mutex.Lock()
	defer gr.mutex.Unlock()

	reqs, ok := gr.gomibakos[key]
	if ok {
		gr.gomibakos[key] = append(reqs, req)
	} else {
		gr.gomibakos[key] = []*http.Request{req}
	}
}

func (gr *gomibakoRepository) Get(key gomibakoKey) []*http.Request {
	gr.mutex.RLock()
	defer gr.mutex.RUnlock()

	reqs, ok := gr.gomibakos[key]
	if ok {
		return reqs
	}

	return []*http.Request{}
}

func main() {

	gr := newGomibakoRepository()
	router := httptreemux.New()
	templates := _templates()

	group := router.UsingContext()
	group.GET("/", func(w http.ResponseWriter, r *http.Request) {
		type Inventry struct {
			Title string
		}
		err := templates["index"].Execute(w, Inventry{Title: "My index page"})
		if err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
	group.POST("/g/-/new", func(w http.ResponseWriter, r *http.Request) {
	})
	group.GET("/g/:gomibakokey/inspect", func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := gomibakoKey(params["gomibakokey"])
		reqs := gr.Get(gomibakoKey)

		type Inventry struct {
			Title    string
			Requests []string
		}

		inv := Inventry{Title: "gomibako: " + string(gomibakoKey), Requests: make([]string, len(reqs))}
		for i, req := range reqs {
			inv.Requests[(len(reqs)-1)-i] = fmt.Sprintf("%v", req)
		}
		err := templates["inspect"].Execute(w, inv)
		if err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
	group.GET("/g/:gomibakokey", func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := gomibakoKey(params["gomibakokey"])

		gr.Add(gomibakoKey, r)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok\n")
	})
	group.POST("/g/:gomibakokey", func(w http.ResponseWriter, r *http.Request) {
	})

	n := negroni.Classic()
	n.UseHandler(router)

	log.Fatal(http.ListenAndServe(":8000", n))
}
