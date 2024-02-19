package generator

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
	"sync"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/room"
)

type pageData struct {
	Title       string
	Version     string
	Description string
	AntiAlias   bool
	XrEnabled   bool
}

type Profile map[string]json.RawMessage

//go:embed html/*
var htmlFs embed.FS

type AppServer struct {
	app        *App
	host, port string
	tls        bool
	certPath   string
	keyPath    string

	webscene *room.WebScene

	serverStarted time.Time
	movelVersion  int
	producerLock  sync.Mutex
}

func (as *AppServer) Serve() error {
	as.serverStarted = time.Now()

	as.webscene = as.app.WebScene
	if as.webscene == nil {
		as.webscene = room.DefaultWebScene()
	}

	htmlData, err := htmlFs.ReadFile("html/server.html")
	if err != nil {
		return err
	}

	pageToServe := pageData{
		Title:       as.app.Name,
		Version:     as.app.Version,
		Description: as.app.Description,
		AntiAlias:   as.webscene.AntiAlias,
		XrEnabled:   as.webscene.XrEnabled,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// Required for sharedMemoryForWorkers to work
		w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Add("Cross-Origin-Resource-Policy", "cross-origin")
		w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")

		t := template.New("")
		_, err := t.Parse(string(htmlData))
		if err != nil {
			panic(err)
		}
		t.Execute(w, pageToServe)
	})

	fSys, err := fs.Sub(htmlFs, "html")
	if err != nil {
		return err
	}

	http.Handle("/js/", http.FileServer(http.FS(fSys)))

	http.HandleFunc("/schema", as.SchemaEndpoint)
	http.HandleFunc("/scene", as.SceneEndpoint)
	http.HandleFunc("/zip", as.ZipEndpoint)
	http.HandleFunc("/started", as.StartedEndpoint)
	http.HandleFunc("/profile", as.ProfileEndpoint)
	http.HandleFunc("/mermaid", as.MermaidEndpoint)
	http.HandleFunc("/producer/", as.ProducerEndpoint)

	hub := room.NewHub(as.webscene, &as.movelVersion)
	go hub.Run()

	http.Handle("/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	}))

	connection := fmt.Sprintf("%s:%s", as.host, as.port)
	if as.tls {
		fmt.Printf("Serving over: https://%s\n", connection)
		return http.ListenAndServeTLS(connection, as.certPath, as.keyPath, nil)

	} else {
		fmt.Printf("Serving over: http://%s\n", connection)
		return http.ListenAndServe(connection, nil)
	}
}

func (as *AppServer) SchemaEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(as.app.Schema())
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func (as *AppServer) ProducerEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Cache-Control", "no-cache")

	// Required for sharedMemoryForWorkers to work
	w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Add("Cross-Origin-Resource-Policy", "cross-origin")
	w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")

	as.producerLock.Lock()
	defer as.producerLock.Unlock()
	// params, _ := url.ParseQuery(r.URL.RawQuery)

	producerToLoad := path.Base(r.URL.Path)

	producer, ok := as.app.Producers[producerToLoad]
	if !ok {
		panic(fmt.Errorf("no producer registered for: %s", producerToLoad))
	}

	artifact := producer.Data()

	bufWr := bufio.NewWriter(w)
	err := artifact.Write(bufWr)
	if err != nil {
		log.Println(err.Error())
		panic(err)
	}
	bufWr.Flush()
}

func (as *AppServer) ApplyProfile(profile Profile) (bool, error) {
	log.Println("applying...")
	params := as.app.Schema()
	// nodes := a.Schema().Nodes

	changed := false
	// for _, p := range params.Nodes {
	// 	paramChanged, err := p.ApplyJsonMessage(profile.Parameters)
	// 	if err != nil {
	// 		return changed, err
	// 	}

	// 	if paramChanged {
	// 		changed = true
	// 	}
	// }

	for key, msg := range profile {
		n := params.Nodes[key]
		_, err := n.parameter.ApplyJsonMessage(msg)
		if err != nil {
			return false, err
		}
		params.Nodes[key] = n
	}

	return changed, nil
}

func (as *AppServer) ProfileEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	as.producerLock.Lock()
	defer as.producerLock.Unlock()

	body, _ := io.ReadAll(r.Body)

	profile := Profile{}
	if err := json.Unmarshal(body, &profile); err != nil {
		panic(err)
	}

	_, err := as.ApplyProfile(profile)
	if err != nil {
		panic(err)
	}

	as.movelVersion++
	w.Write([]byte("{}"))
}

func (as *AppServer) StartedEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	time := as.serverStarted.Format("2006-01-02 15:04:05")
	w.Write([]byte(fmt.Sprintf("{ \"time\": \"%s\" }", time)))
}

func (as *AppServer) MermaidEndpoint(w http.ResponseWriter, r *http.Request) {
	err := WriteMermaid(*as.app, w)
	if err != nil {
		log.Println(err.Error())
	}
}

func (as *AppServer) ZipEndpoint(w http.ResponseWriter, r *http.Request) {
	err := as.app.WriteZip(w)
	w.Header().Add("Content-Type", "application/zip")
	if err != nil {
		panic(err)
	}
}

func (as *AppServer) SceneEndpoint(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(as.webscene)
	if err != nil {
		panic(err)
	}
	w.Write(data)
}
