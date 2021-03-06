package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hoisie/mustache"
	"github.com/martingartonft/timemachine/api"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

func main() {

	f, err := os.Create("/tmp/cpuprof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	index, err := api.NewGitContentAPI()
	if err != nil {
		panic(err)
	}

	ah := apiHandlers{index}

	m := mux.NewRouter()

	logEndpointsAndRegisterHandlers(m, "/content/recent", ah.recentHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/count", ah.countHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/{uuid}/versions/", ah.versionsHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/{uuid}/versions/{ver}", ah.versionHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/{uuid}", ah.uuidReadHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/{uuid}", ah.idWriteHandler, "PUT")
	logEndpointsAndRegisterHandlers(m, "/content/", ah.dropHandler, "DELETE")
	logEndpointsAndRegisterHandlers(m, "/content/", ah.dumpAllHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/article/{uuid}", ah.getArticlePageHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/article/{uuid}/versions/{version}", ah.getVersionedArticlePageHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/", ah.indexHandler, "GET")

	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, m))

	go func() {
		port := "8082"
		fmt.Printf("listening on port: %s ...", port)
		err = http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Printf("web stuff failed: %v\n", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// wait for ctrl-c
	<-c
	println("exiting")
	index.Close()

	f, err = os.Create("/tmp/memprof")
	if err != nil {
		panic(err)
	}

	pprof.WriteHeapProfile(f)
	f.Close()

	return
}

func logEndpointsAndRegisterHandlers(m *mux.Router, route string, handlerMethod func(w http.ResponseWriter, r *http.Request), httpMethod string) {
	log.Printf("Registering %[1]s %[2]s \n", httpMethod, route)
	m.HandleFunc(route, handlerMethod).Methods(httpMethod)
}

type apiHandlers struct {
	index api.ContentAPI
}

func (ah *apiHandlers) getArticlePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	found, art := ah.index.ByUUID(uuid)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("content with id %s was not found\n", uuid)))
		return
	}

	data := make(map[string]interface{})
	data["art"] = art
	data["versions"] = ah.index.Versions(uuid)
	renderedHTML := mustache.RenderFile("./static/article.html", data)
	io.Copy(w, strings.NewReader(renderedHTML))
}

func (ah *apiHandlers) getVersionedArticlePageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	version := vars["version"]

	found, art := ah.index.Version(uuid, version)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("content with id %s was not found\n", uuid)))
		return
	}

	data := make(map[string]interface{})
	data["art"] = art
	data["versions"] = ah.index.Versions(uuid)
	renderedHTML := mustache.RenderFile("./static/article.html", data)

	io.Copy(w, strings.NewReader(renderedHTML))
}

func (ah *apiHandlers) versionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]
	ver := vars["ver"]

	found, art := ah.index.Version(uuid, ver)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("content with uuid %s and version %s was not found\n", uuid, ver)))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(art)

}

func (ah *apiHandlers) versionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	versions := ah.index.Versions(uuid)
	if len(versions) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("content with uuid %s was not found\n", uuid)))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(versions)
}

func (ah *apiHandlers) uuidAndDateTimeReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["uuid"]
	timestamp := r.URL.Query().Get("atTime")

	timestampAsDateTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		panic(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("Error parsing timestamp: %s \n", timestamp)))
		return
	}

	found, art := ah.index.ByUUIDAndDate(id, timestampAsDateTime)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("content with id %s was not found\n", id)))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(art)

}

func (ah *apiHandlers) uuidReadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["uuid"]

	found, art := ah.index.ByUUID(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("content with id %s was not found\n", id)))
		return
	}
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(art)
}

func (ah *apiHandlers) idWriteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	var c api.Content
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if c.UUID != uuid {
		http.Error(w, "id does not match", http.StatusBadRequest)
		return
	}

	err = ah.index.Write(c)
	if err != nil {
		http.Error(w, fmt.Sprintf("write failed:\n%v\n", err), http.StatusInternalServerError)
		return
	}
}

func (ah *apiHandlers) dropHandler(w http.ResponseWriter, r *http.Request) {
	ah.index.Drop()
}

func (ah *apiHandlers) recentHandler(w http.ResponseWriter, r *http.Request) {
	count := 20
	r.ParseForm()
	max := r.Form["max"]
	if len(max) == 1 {
		i, err := strconv.Atoi(max[0])
		if err == nil {
			count = i
		}
	}
	stop := make(chan struct{})
	defer close(stop)
	cont, err := ah.index.Recent(stop, count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	first := true
	enc := json.NewEncoder(w)
	fmt.Fprint(w, "[\n")
	for c := range cont {
		if first {
			first = false
		} else {
			fmt.Fprint(w, ",")
		}
		err := enc.Encode(c)
		if err != nil {
			log.Printf("error writing json to response: %v\n", err)
			return
		}
	}
	fmt.Fprintf(w, "]")
}

func (ah *apiHandlers) dumpAllHandler(w http.ResponseWriter, r *http.Request) {
	first := true
	enc := json.NewEncoder(w)
	allContent, err := ah.index.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	for _, content := range allContent {
		if first {
			fmt.Fprint(w, "[\n")
			first = false
		} else {
			fmt.Fprint(w, ",\n")
		}
		enc.Encode(content)
	}
	fmt.Fprintf(w, "]")
}

func (ah *apiHandlers) indexHandler(w http.ResponseWriter, r *http.Request) {
	allContent, err := ah.index.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vars := make(map[string]interface{})
	vars["content"] = allContent
	fmt.Printf("CONT:%v\n", vars)
	w.Header().Add("Content-Type", "text/html")
	s := mustache.RenderFile("static/index.html", vars)
	io.Copy(w, strings.NewReader(s))
}

func (ah *apiHandlers) countHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d", ah.index.Count())
}
