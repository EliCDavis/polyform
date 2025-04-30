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
	"runtime/debug"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/endpoint"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
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
	Title         string
	Version       string
	Description   string
	AntiAlias     bool
	XrEnabled     bool
	ExampleGraphs []string
}

//go:embed html/*
var htmlFs embed.FS

type AppServer struct {
	app              *App
	host, port       string
	tls              bool
	certPath         string
	keyPath          string
	launchWebbrowser bool

	autosave   bool
	configPath string

	webscene *schema.WebScene

	serverStarted time.Time

	clientConfig *room.ClientConfig
}

func (as *AppServer) Handler(indexFile string) (*http.ServeMux, error) {
	as.serverStarted = time.Now()

	as.webscene = as.app.WebScene
	if as.webscene == nil {
		as.webscene = room.DefaultWebScene()
	}

	mux := http.NewServeMux()

	htmlData, err := htmlFs.ReadFile("html/server.html")
	if err != nil {
		return nil, err
	}

	htmlTemplate := template.New("")
	_, err = htmlTemplate.Parse(string(htmlData))
	if err != nil {
		return nil, fmt.Errorf("unable to interpret html template: %w", err)
	}

	mux.HandleFunc(indexFile, func(w http.ResponseWriter, r *http.Request) {
		pageToServe := pageData{
			Title:         as.app.Name,
			Version:       as.app.Version,
			Description:   as.app.Description,
			AntiAlias:     as.webscene.AntiAlias,
			XrEnabled:     as.webscene.XrEnabled,
			ExampleGraphs: allExamples(),
		}

		// Required for sharedMemoryForWorkers to work
		w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Add("Cross-Origin-Resource-Policy", "cross-origin")
		w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")

		err := htmlTemplate.Execute(w, pageToServe)
		if err != nil {
			panic(err)
		}
	})

	fSys, err := fs.Sub(htmlFs, "html")
	if err != nil {
		return nil, err
	}

	fs := http.FileServer(http.FS(fSys))
	mux.Handle("/js/", fs)

	var graphSaver *GraphSaver
	if as.autosave {
		graphSaver = &GraphSaver{
			app:      as.app,
			savePath: as.configPath,
		}
	}

	mux.HandleFunc("/schema", as.SchemaEndpoint)
	mux.Handle("/scene", endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[*schema.WebScene]{
				ResponseWriter: endpoint.JsonResponseWriter[*schema.WebScene]{},
				Handler: func(r *http.Request) (*schema.WebScene, error) {
					return as.webscene, nil
				},
			},
		},
	})
	mux.HandleFunc("/zip/", as.ZipEndpoint)
	mux.HandleFunc("/node-types", NodeTypesEndpoint)
	mux.Handle("/node", nodeEndpoint(as.app.graphInstance, graphSaver))
	mux.Handle("/node/connection", nodeConnectionEndpoint(as.app.graphInstance, graphSaver))
	mux.Handle("/parameter/value/", parameterValueEndpoint(as.app.graphInstance, graphSaver))
	mux.Handle("/parameter/name/", parameterNameEndpoint(as.app.graphInstance, graphSaver))
	mux.Handle("/parameter/description/", parameterDescriptionEndpoint(as.app.graphInstance, graphSaver))
	mux.Handle("/new-graph", newGraphEndpoint(as.app))
	mux.Handle("/load-example", exampleGraphEndpoint(as.app))
	mux.Handle("/graph", graphEndpoint(as.app))
	mux.Handle("/graph/metadata/", graphMetadataEndpoint(as.app.graphInstance, graphSaver))
	mux.HandleFunc("/started", as.StartedEndpoint)
	mux.HandleFunc("/mermaid", as.MermaidEndpoint)
	mux.HandleFunc("/swagger", as.SwaggerEndpoint)
	mux.HandleFunc("/producer/value/", as.ProducerEndpoint)
	mux.Handle("/producer/name/", producerNameEndpoint(as.app.graphInstance, graphSaver))
	mux.HandleFunc("/manifest/", as.ManifestEndpoint)

	hub := room.NewHub(as.webscene, as.app.graphInstance)
	go hub.Run()

	mux.Handle("/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := as.clientConfig
		if conf == nil {
			conf = room.DefaultClientConfig()
		}
		hub.ServeWs(w, r, conf)
	}))

	return mux, nil
}

func (as *AppServer) SchemaEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(as.app.graphInstance.Schema())
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func NodeTypesEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(graph.BuildSchemaForAllNodeTypes(types))
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

	// params, _ := url.ParseQuery(r.URL.RawQuery)
	err := as.writeProducerDataToRequest(path.Base(path.Dir(r.URL.Path)), path.Base(r.URL.Path), w)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
	}
}

func (as *AppServer) writeProducerDataToRequest(producerToLoad, file string, w http.ResponseWriter) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	manifest := as.app.graphInstance.Manifest(producerToLoad)
	artifact := manifest.Entries[file].Artifact

	w.Header().Set("Content-Type", artifact.Mime())

	bufWr := bufio.NewWriter(w)
	err = artifact.Write(bufWr)
	if err != nil {
		return
	}
	return bufWr.Flush()
}

func (as *AppServer) StartedEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	time := as.serverStarted.Format("2006-01-02 15:04:05")
	w.Write([]byte(fmt.Sprintf("{ \"time\": \"%s\", \"modelVersion\": %d }", time, as.app.graphInstance.ModelVersion())))
}

func (as *AppServer) MermaidEndpoint(w http.ResponseWriter, r *http.Request) {
	err := WriteMermaid(*as.app, w)
	if err != nil {
		log.Println(err.Error())
	}
}

func (as *AppServer) SwaggerEndpoint(w http.ResponseWriter, r *http.Request) {
	err := as.app.WriteSwagger(w)
	if err != nil {
		log.Println(err.Error())
	}
}

func (as *AppServer) SceneEndpoint(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(as.webscene)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
