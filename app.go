package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/martingartonft/timemachine/api"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
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
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, m))

	logEndpointsAndRegisterHandlers(m, "/content/recent", ah.recentHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/count", ah.countHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/{uuid}", ah.uuidReadHandler, "GET")
	logEndpointsAndRegisterHandlers(m, "/content/{uuid}", ah.idWriteHandler, "PUT")
	logEndpointsAndRegisterHandlers(m, "/content/", ah.dropHandler, "DELETE")
	logEndpointsAndRegisterHandlers(m, "/content/", ah.dumpAll, "GET")

	//m.HandleFunc("/content/recent", ah.recentHandler).Methods("GET")
	//m.HandleFunc("/content/count", ah.countHandler).Methods("GET")
	//m.HandleFunc("/content/{uuid}", ah.uuidReadHandler).Methods("GET")
	//m.HandleFunc("/content/{uuid}", ah.idWriteHandler).Methods("PUT")
	//m.HandleFunc("/content/", ah.dropHandler).Methods("DELETE")
	//m.HandleFunc("/content/", ah.dumpAll).Methods("GET")

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

func (ah *apiHandlers) dumpAll(w http.ResponseWriter, r *http.Request) {
	first := true
	enc := json.NewEncoder(w)
	stop := make(chan struct{})
	defer close(stop)
	allContent, err := ah.index.All(stop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for content := range allContent {
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

func (ah *apiHandlers) countHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%d", ah.index.Count())
}
