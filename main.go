package main

import (
	"bytes"
	"flag"
	"log"
	"net/http"

	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"sort"
	"strconv"
	"time"

	gomibako "github.com/hakobe/gomibako/lib"

	"github.com/dimfeld/httptreemux"
	"github.com/unrolled/secure"
	"github.com/urfave/negroni"
)

func main() {
	port := flag.Uint64("port", 8000, "Binding TCP port")
	flag.Parse()

	gr := gomibako.NewGomibakoRepository()
	go gr.RunTruncater()
	go gr.RunBroker()

	handler := NewServerHandler(gr)
	log.Fatal(http.ListenAndServe(":"+strconv.FormatUint(*port, 10), handler))
}

type ViewableHeaderPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ViewableHeaders []*ViewableHeaderPair

func (hs ViewableHeaders) Len() int      { return len(hs) }
func (hs ViewableHeaders) Swap(i, j int) { hs[i], hs[j] = hs[j], hs[i] }
func (hs ViewableHeaders) Less(i, j int) bool {
	if hs[i].Key == hs[j].Key {
		return hs[i].Value < hs[j].Value
	} else {
		return hs[i].Key < hs[j].Key
	}
}

type ViewableGomibakoRequest struct {
	Timestamp     int64           `json:"timestamp"`
	Method        string          `json:"method"`
	URL           string          `json:"url"`
	Headers       ViewableHeaders `json:"headers"`
	Body          string          `json:"body"`
	ContentLength string          `json:"contentLength"`
}

func NewViewableGomibakoRequest(greq *gomibako.GomibakoRequest) *ViewableGomibakoRequest {
	var viewableHeaders ViewableHeaders
	for k, vs := range greq.Headers {
		for _, v := range vs {
			viewableHeaders = append(viewableHeaders, &ViewableHeaderPair{k, v})
		}
	}
	sort.Sort(viewableHeaders)
	bodyRunes := bytes.Runes(greq.Body)
	if len(bodyRunes) > 1000*1000 {
		bodyRunes = bodyRunes[0:(1000 * 1000)]
	}
	return &ViewableGomibakoRequest{
		Timestamp:     greq.Timestamp.UnixNano(),
		Method:        greq.Method,
		URL:           greq.URL.String(),
		Headers:       viewableHeaders,
		Body:          string(bodyRunes),
		ContentLength: strconv.Itoa(greq.ContentLength),
	}
}

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

func NewServerHandler(gr *gomibako.GomibakoRepository) http.Handler {
	router := httptreemux.New()
	templates := _templates()

	group := router.UsingContext()
	group.GET("/", func(w http.ResponseWriter, r *http.Request) {
		type Inventry struct {
			Title string
		}
		err := templates["index"].Execute(w, Inventry{Title: ""})
		if err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
	group.POST("/g/-/new", func(w http.ResponseWriter, r *http.Request) {
		g, err := gr.AddGomibako()
		if err != nil {
			http.Error(w, "gomibako generation error: "+err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, "/g/"+string(g.Key)+"/inspect", 302)
	})
	group.GET("/g/:gomibakokey/inspect", func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := gomibako.GomibakoKey(params["gomibakokey"])

		_, err := gr.Get(gomibakoKey)
		if err != nil {
			http.Error(w, "no gomibako found", http.StatusNotFound)
			return
		}

		type Inventry struct {
			Title string
		}

		err = templates["inspect"].Execute(w, Inventry{Title: string(gomibakoKey)})
		if err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
	})
	group.GET("/g/:gomibakokey/reqevents", func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := gomibako.GomibakoKey(params["gomibakokey"])
		g, ch, err := gr.GetWithCh(gomibakoKey)
		if err != nil {
			http.Error(w, "no gomibako found", http.StatusNotFound)
			return
		}
		defer gr.Release(g.Key, ch)

		fw, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		cw, ok := w.(http.CloseNotifier)
		if !ok {
			http.Error(w, "Close notifying unsupported!", http.StatusInternalServerError)
			return
		}

		cn := cw.CloseNotify()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		greqs := g.Requests()
		for _, greq := range greqs {
			j, err := json.Marshal(NewViewableGomibakoRequest(greq))
			if err != nil {
				http.Error(w, "Failed to create json", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", j)
		}
		fw.Flush()

		for {
			select {
			case gomibakoReq, ok := <-ch:
				if !ok {
					return
				}
				j, err := json.Marshal(NewViewableGomibakoRequest(gomibakoReq))
				if err != nil {
					http.Error(w, "Failed to create json", http.StatusInternalServerError)
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", j)
				fw.Flush()
			case _ = <-cn:
				return
			}
		}
	})
	recordReq := func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := gomibako.GomibakoKey(params["gomibakokey"])
		_, err := gr.Get(gomibakoKey)
		if err != nil {
			http.Error(w, "no gomibako found", http.StatusNotFound)
			return
		}

		reader := http.MaxBytesReader(w, r.Body, 3*1000*1000)
		defer reader.Close()
		bodyBytes, err := ioutil.ReadAll(reader)
		if err != nil {
			http.Error(w, "failed to load body", http.StatusBadRequest)
			return
		}

		greq := &gomibako.GomibakoRequest{
			Key:           gomibakoKey,
			Timestamp:     time.Now(),
			Method:        r.Method,
			URL:           r.URL,
			Headers:       r.Header,
			Body:          bodyBytes,
			ContentLength: len(bodyBytes),
		}
		gr.AddRequest(greq)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok\n")
	}
	group.GET("/g/:gomibakokey", recordReq)
	group.POST("/g/:gomibakokey", recordReq)

	n := negroni.Classic()
	secureMiddleware := secure.New(secure.Options{
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'",
	})
	n.Use(negroni.HandlerFunc(secureMiddleware.HandlerFuncWithNext))
	n.Use(negroni.NewStatic(http.Dir("static")))
	n.UseHandler(router)

	return n
}
