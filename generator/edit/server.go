package edit

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
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/serialize"
	"github.com/EliCDavis/polyform/generator/variable"
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
	Title             string
	Version           string
	Description       string
	AntiAlias         bool
	XrEnabled         bool
	ShowNewGraphPopup bool
	ExampleGraphs     []string
}

//go:embed html/*
var htmlFs embed.FS

type Server struct {
	Graph                   *graph.Instance
	Host, Port              string
	Tls                     bool
	CertPath                string
	KeyPath                 string
	LaunchWebbrowser        bool
	VariableFactory         func(string) (variable.Variable, error)
	NodeOutputSerialization *serialize.TypeSwitch[manifest.Entry]

	Autosave   bool
	ConfigPath string

	Webscene *schema.WebScene

	ClientConfig *room.ClientConfig

	serverStarted     time.Time
	showNewGraphPopup bool
}

func (as *Server) Handler(indexFile string) (*http.ServeMux, error) {
	as.serverStarted = time.Now()
	as.showNewGraphPopup = as.ConfigPath == ""

	if as.Webscene == nil {
		as.Webscene = room.DefaultWebScene()
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
		title := as.Graph.GetName()
		description := as.Graph.GetDescription()

		if description == "" && title == "" {
			title = "Polyform"
			description = "Immutable mesh processing pipelines"
		}

		pageToServe := pageData{
			Title:             title,
			Version:           as.Graph.GetVersion(),
			Description:       description,
			AntiAlias:         as.Webscene.AntiAlias,
			XrEnabled:         as.Webscene.XrEnabled,
			ShowNewGraphPopup: as.showNewGraphPopup,
			ExampleGraphs:     allExamples(),
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
	mux.Handle("/icons/", fs)

	var graphSaver *GraphSaver
	if as.Autosave {
		graphSaver = &GraphSaver{
			graph:    as.Graph,
			savePath: as.ConfigPath,
		}
	}

	mux.HandleFunc("/schema", as.SchemaEndpoint)
	mux.Handle("/scene", endpoint.Handler{
		Methods: map[string]endpoint.Method{
			http.MethodGet: endpoint.ResponseMethod[*schema.WebScene]{
				ResponseWriter: endpoint.JsonResponseWriter[*schema.WebScene]{},
				Handler: func(r *http.Request) (*schema.WebScene, error) {
					return as.Webscene, nil
				},
			},
		},
	})
	mux.HandleFunc("/zip/", as.ZipEndpoint)
	mux.Handle("/node-types", nodeTypesEndpoint(as.Graph, as.NodeOutputSerialization))
	mux.Handle("/node", nodeEndpoint(as.Graph, graphSaver))
	mux.Handle("/node/connection", nodeConnectionEndpoint(as.Graph, graphSaver))
	mux.HandleFunc(nodeOutputEndpointPath, as.NodeOutputEndpoint)
	mux.Handle("/parameter/value/", parameterValueEndpoint(as.Graph, graphSaver))
	mux.Handle("/parameter/name/", parameterNameEndpoint(as.Graph, graphSaver))
	mux.Handle("/parameter/description/", parameterDescriptionEndpoint(as.Graph, graphSaver))

	mux.Handle("/profile", profileEndpoint(as.Graph, graphSaver))
	mux.Handle("/profile/apply", applyProfileEndpoint(as.Graph, graphSaver))
	mux.Handle("/profile/rename", renameProfileEndpoint(as.Graph, graphSaver))
	mux.Handle("/profile/overwrite", overwriteProfileEndpoint(as.Graph, graphSaver))

	mux.Handle("/new-graph", newGraphEndpoint(as))
	mux.Handle("/load-example", exampleGraphEndpoint(as))
	mux.Handle("/graph", graphEndpoint(as))
	mux.Handle("/graph/execution-report", executionReportEndpoint(as))
	mux.Handle("/graph/metadata/", graphMetadataEndpoint(as.Graph, graphSaver))
	mux.HandleFunc("/started", as.StartedEndpoint)
	mux.HandleFunc("/mermaid", as.MermaidEndpoint)
	mux.HandleFunc("/swagger", as.SwaggerEndpoint)
	mux.HandleFunc("/producer/value/", as.ProducerEndpoint)
	mux.Handle("/producer/name/", producerNameEndpoint(as.Graph, graphSaver))
	mux.HandleFunc("/manifest/", as.ManifestEndpoint)
	mux.Handle(variableInstanceEndpointPath, variableInstanceEndpoint(as, as.Graph, graphSaver))
	mux.Handle(variableValueEndpointPath, variableValueEndpoint(as.Graph, graphSaver))
	mux.Handle(variableNameDescriptionEndpointPath, variableInfoEndpoint(as.Graph, graphSaver))

	hub := room.NewHub(as.Webscene, as.Graph)
	go hub.Run()

	mux.Handle("/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conf := as.ClientConfig
		if conf == nil {
			conf = room.DefaultClientConfig()
		}
		hub.ServeWs(w, r, conf)
	}))

	return mux, nil
}

func (as *Server) SchemaEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(as.Graph.Schema())
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func (as *Server) ProducerEndpoint(w http.ResponseWriter, r *http.Request) {
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

func (as *Server) writeProducerDataToRequest(producerToLoad, file string, w http.ResponseWriter) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	manifest := as.Graph.Manifest(producerToLoad)
	artifact := manifest.Entries[file].Artifact

	w.Header().Set("Content-Type", artifact.Mime())

	bufWr := bufio.NewWriter(w)
	err = artifact.Write(bufWr)
	if err != nil {
		return
	}
	return bufWr.Flush()
}

func (as *Server) StartedEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	time := as.serverStarted.Format("2006-01-02 15:04:05")
	fmt.Fprintf(w, "{ \"time\": \"%s\", \"modelVersion\": %d }", time, as.Graph.ModelVersion())
}

func (as *Server) MermaidEndpoint(w http.ResponseWriter, r *http.Request) {
	err := graph.WriteMermaid(as.Graph, w)
	if err != nil {
		log.Println(err.Error())
	}
}

func (as *Server) SwaggerEndpoint(w http.ResponseWriter, r *http.Request) {
	err := graph.WriteSwagger(as.Graph, w)
	if err != nil {
		log.Println(err.Error())
	}
}

func (as *Server) SceneEndpoint(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(as.Webscene)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
