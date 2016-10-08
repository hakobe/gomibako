package main

import (
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"time"

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

type ViewableHeaderPair struct {
	Key   string
	Value string
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
	Timestamp     string
	Method        string
	URL           string
	Headers       ViewableHeaders
	Body          string
	ContentLength string
}

func NewViewableGomibakoRequest(greq *GomibakoRequest) *ViewableGomibakoRequest {
	var viewableHeaders ViewableHeaders
	for k, vs := range greq.headers {
		for _, v := range vs {
			viewableHeaders = append(viewableHeaders, &ViewableHeaderPair{k, v})
		}
	}
	sort.Sort(viewableHeaders)
	return &ViewableGomibakoRequest{
		Timestamp:     greq.timestamp.String(),
		Method:        greq.method,
		URL:           greq.url.String(),
		Headers:       viewableHeaders,
		Body:          string(greq.body),
		ContentLength: strconv.FormatInt(greq.contentLength, 10),
	}
}

func NewServerHandler(gr *GomibakoRepository) http.Handler {
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
		g, err := gr.AddGomibako()
		if err != nil {
			http.Error(w, "gomibako generation error: "+err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, "/g/"+string(g.key)+"/inspect", 302)
	})
	group.GET("/g/:gomibakokey/inspect", func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := GomibakoKey(params["gomibakokey"])
		g, err := gr.Get(gomibakoKey)
		if err != nil {
			http.Error(w, "no gomibako found", http.StatusNotFound)
			return
		}
		var requests []*ViewableGomibakoRequest
		for r := g.requests.Back(); r != nil; r = r.Prev() {
			requests = append(requests, NewViewableGomibakoRequest(r.Value.(*GomibakoRequest)))
		}

		type Inventry struct {
			Title    string
			Requests []*ViewableGomibakoRequest
		}

		inv := Inventry{Title: "gomibako: " + string(gomibakoKey), Requests: requests}
		err = templates["inspect"].Execute(w, inv)
		if err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})
	recordReq := func(w http.ResponseWriter, r *http.Request) {
		params := httptreemux.ContextParams(r.Context())
		gomibakoKey := GomibakoKey(params["gomibakokey"])
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

		greq := &GomibakoRequest{
			timestamp: time.Now(),
			method:    r.Method,
			url:       r.URL,
			headers:   r.Header,
			body:      bodyBytes,
		}
		gr.AddRequest(gomibakoKey, greq)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "ok\n")
	}
	group.GET("/g/:gomibakokey", recordReq)
	group.POST("/g/:gomibakokey", recordReq)

	n := negroni.Classic()
	n.UseHandler(router)

	return n
}
