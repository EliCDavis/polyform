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

func writeJSONError(out io.Writer, err error) error {
	var d struct {
		Error string `json:"error"`
	} = struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, err = out.Write(data)
	return err
}

type pageData struct {
	Title       string
	Version     string
	Description string
	AntiAlias   bool
	XrEnabled   bool
}

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
	movelVersion  uint32
	producerLock  sync.Mutex

	clientConfig *room.ClientConfig
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

	fs := http.FileServer(http.FS(fSys))
	http.Handle("/js/", fs)
	http.Handle("/css/", fs)

	http.HandleFunc("/schema", as.SchemaEndpoint)
	http.HandleFunc("/scene", as.SceneEndpoint)
	http.HandleFunc("/zip", as.ZipEndpoint)
	http.HandleFunc("/started", as.StartedEndpoint)
	http.HandleFunc("/profile/", as.ProfileEndpoint)
	http.HandleFunc("/mermaid", as.MermaidEndpoint)
	http.HandleFunc("/producer/", as.ProducerEndpoint)

	hub := room.NewHub(as.webscene, &as.movelVersion)
	go hub.Run()

	http.Handle("/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := as.clientConfig
		if conf == nil {
			conf = room.DefaultClientConfig()
		}
		hub.ServeWs(w, r, conf)
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

func (as *AppServer) ApplyMessage(key string, msg []byte) (bool, error) {
	log.Println("applying...")
	params := as.app.Schema()
	// nodes := a.Schema().Nodes

	// for _, p := range params.Nodes {
	// 	paramChanged, err := p.ApplyJsonMessage(profile.Parameters)
	// 	if err != nil {
	// 		return changed, err
	// 	}

	// 	if paramChanged {
	// 		changed = true
	// 	}
	// }

	n := params.Nodes[key]
	changed, err := n.parameter.ApplyMessage(msg)
	params.Nodes[key] = n

	return changed, err
}

func (as *AppServer) ProfileEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	as.producerLock.Lock()
	defer as.producerLock.Unlock()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	_, err = as.ApplyMessage(path.Base(r.URL.Path), body)
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
