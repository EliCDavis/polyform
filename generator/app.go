package generator

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/room"
	"github.com/EliCDavis/polyform/nodes"
)

type producerCache map[string]Artifact

func (pc producerCache) Lookup(producer string) Artifact {
	if pc == nil {
		return nil
	}

	if artifact, ok := pc[producer]; ok {
		return artifact
	}

	return nil
}

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *room.WebScene
	Authors     []Author
	Producers   map[string]nodes.NodeOutput[Artifact]

	// Runtime data
	producerCache producerCache
}

type pageData struct {
	Title       string
	Version     string
	Description string
	Scripting   string
}

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

func writeProducersToZip(path string, producers map[string]nodes.NodeOutput[Artifact], zw *zip.Writer, cache producerCache) error {
	if producers == nil {
		panic("can't write nil producers")
	}

	if zw == nil {
		panic("can't write to nil zip writer")
	}

	for file, producer := range producers {
		filePath := path + file
		f, err := zw.Create(filePath)
		if err != nil {
			return err
		}

		artifact := cache.Lookup(file)
		if artifact == nil {
			artifact = producer.Data()
			// artifact, err = producer.Data()
			// if err != nil {
			// return err
			// }
		}

		err = artifact.Write(f)
		if err != nil {
			return err
		}
		// log.Printf("wrote %s", filePath)
	}

	return nil
}

func (a App) getParameters() []Parameter {
	if a.Producers == nil {
		return nil
	}

	parameterSet := make(map[Parameter]struct{})
	for _, n := range a.Producers {
		params := recurseDependenciesType[Parameter](n.Node())
		for _, p := range params {
			parameterSet[p] = struct{}{}
		}
	}

	uniqueParams := make([]Parameter, 0, len(parameterSet))
	for p := range parameterSet {
		uniqueParams = append(uniqueParams, p)
	}
	return uniqueParams
}

func recurseDependenciesType[T any](dependent nodes.Dependent) []T {
	allDependencies := make([]T, 0)
	for _, dep := range dependent.Dependencies() {
		subDependent, ok := dep.(nodes.Dependent)
		if ok {
			subDependencies := recurseDependenciesType[T](subDependent)
			allDependencies = append(allDependencies, subDependencies...)
		}

		ofT, ok := dep.(T)
		if ok {
			allDependencies = append(allDependencies, ofT)
		}
	}

	return allDependencies
}

func (a App) initialize(set *flag.FlagSet) {
	for _, p := range a.getParameters() {
		p.initializeForCLI(set)
	}
}

func (a App) WriteZip(out io.Writer) error {
	z := zip.NewWriter(out)

	err := writeProducersToZip("", a.Producers, z, a.producerCache)
	if err != nil {
		return err
	}

	return z.Close()
}

func (a App) WriteMermaid(out io.Writer) error {

	schema := a.Schema()

	fmt.Fprintf(out, "---\ntitle: %s\n---\n\nflowchart LR\n", a.Name)

	for id, n := range schema.Nodes {

		if len(n.Dependencies) > 0 {
			fmt.Fprintf(out, "\tsubgraph %s[%s]\n\tdirection TB\n", id, n.Name)
			fmt.Fprintf(out, "\tsubgraph %s-I[%s]\n\tdirection TB\n", id, "Input")
		} else {
			fmt.Fprintf(out, "\t%s[%s]\n", id, n.Name)
		}

		for depIndex, dep := range n.Dependencies {
			fmt.Fprintf(out, "\t%s-%d([%s])\n", id, depIndex, dep.Name)
		}

		if len(n.Dependencies) > 0 {
			fmt.Fprint(out, "\tend\n")
			fmt.Fprint(out, "\tend\n")
		}
	}

	for id, n := range schema.Nodes {
		for depIndex, d := range n.Dependencies {
			fmt.Fprintf(out, "\t%s --> %s-%d\n", d.DependencyID, id, depIndex)
		}
	}

	return nil
}

//go:embed html/*
var htmlFs embed.FS

//go:embed cli.tmpl
var cliTemplate string

type appCLI struct {
	Name        string
	Version     string
	Description string
	Authors     []Author
	Commands    []*cliCommand
}

func buildSchemaForNode(dependency nodes.Node, currentSchema map[string]NodeSchema, idMapping map[nodes.Node]string) {

	if _, ok := idMapping[dependency]; ok {
		return
	}

	schema := NodeSchema{
		Name:         "Unamed",
		Dependencies: make([]NodeDependencySchema, 0),
		Version:      dependency.Version(),
	}

	for _, subDependency := range dependency.Dependencies() {
		buildSchemaForNode(subDependency.Dependency(), currentSchema, idMapping)
		schema.Dependencies = append(schema.Dependencies, NodeDependencySchema{
			DependencyID: idMapping[subDependency.Dependency()],
			Name:         subDependency.Name(),
		})
	}

	if param, ok := dependency.(Parameter); ok {
		schema.Name = param.DisplayName()
	} else {
		named, ok := dependency.(nodes.Named)
		if ok {
			schema.Name = named.Name()
		}
	}

	id := fmt.Sprintf("Node-%d", len(currentSchema))
	idMapping[dependency] = id
	currentSchema[id] = schema
}

func (a *App) Schema() AppSchema {
	schema := AppSchema{
		Producers: make([]string, 0, len(a.Producers)),
	}

	nodeSchema := make(map[string]NodeSchema)
	tempIDData := make(map[nodes.Node]string)

	for key, producer := range a.Producers {
		schema.Producers = append(schema.Producers, key)
		buildSchemaForNode(producer.Node(), nodeSchema, tempIDData)

		node := nodeSchema[tempIDData[producer.Node()]]
		node.Name = key
		nodeSchema[tempIDData[producer.Node()]] = node
	}

	schema.Nodes = nodeSchema

	return schema
}

func (a *App) ApplyProfile(profile Profile) (bool, error) {

	params := a.getParameters()

	changed := false

	for _, p := range params {
		paramChanged, err := p.ApplyJsonMessage(profile.Parameters)
		if err != nil {
			return changed, err
		}

		if paramChanged {
			changed = true
		}
	}

	return changed, nil
}

func (a *App) Serve(host, port string) error {

	a.producerCache = make(map[string]Artifact)
	producerLock := &sync.Mutex{}

	serverStarted := time.Now()

	webscene := a.WebScene
	if webscene == nil {
		webscene = room.DefaultWebScene()
	}

	movelVersion := 0

	htmlData, err := htmlFs.ReadFile("html/server.html")
	if err != nil {
		return err
	}

	javascriptData, err := htmlFs.ReadFile("html/index.js")
	if err != nil {
		return err
	}

	pageToServe := pageData{
		Title:       a.Name,
		Version:     a.Version,
		Description: a.Description,
		Scripting:   " <script type=\"module\">" + string(javascriptData) + "</script>",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// d, err := os.ReadFile("generator/server.html")
		// if err != nil {
		// 	panic(err)
		// }

		t := template.New("")
		// _, err = t.Parse(string(d))
		_, err := t.Parse(string(htmlData))
		if err != nil {
			panic(err)
		}
		t.Execute(w, pageToServe)
		// w.Write(indexData)
	})

	http.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(a.Schema())
		if err != nil {
			panic(err)
		}
		w.Write(data)
	})

	http.HandleFunc("/scene", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(webscene)
		if err != nil {
			panic(err)
		}
		w.Write(data)
	})

	http.HandleFunc("/started", func(w http.ResponseWriter, r *http.Request) {
		time := serverStarted.Format("2006-01-02 15:04:05")
		w.Write([]byte(fmt.Sprintf("{ \"time\": \"%s\" }", time)))
	})

	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		producerLock.Lock()
		defer producerLock.Unlock()

		body, _ := io.ReadAll(r.Body)

		profile := Profile{}
		if err := json.Unmarshal(body, &profile); err != nil {
			panic(err)
		}

		_, err := a.ApplyProfile(profile)
		if err != nil {
			panic(err)
		}

		movelVersion++
		w.Write([]byte("{}"))
	})

	http.HandleFunc("/producer/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")

		producerLock.Lock()
		defer producerLock.Unlock()
		// params, _ := url.ParseQuery(r.URL.RawQuery)

		producerToLoad := path.Base(r.URL.Path)

		producer, ok := a.Producers[producerToLoad]
		if !ok {
			panic(fmt.Errorf("no producer registered for: %s", producerToLoad))
		}

		if artifact, ok := a.producerCache[producerToLoad]; ok && artifact != nil {
			artifact.Write(w)
			return
		}

		artifact := producer.Data()

		// artifact, err := producer.Data()
		// if err != nil {
		// 	log.Printf(err.Error())
		// 	w.Header().Add("Content-Type", "application/json")
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	err := writeJSONError(w, err)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	return
		// }

		err = artifact.Write(w)
		if err != nil {
			panic(err)
		}

		a.producerCache[producerToLoad] = artifact
	})

	http.HandleFunc("/zip", func(w http.ResponseWriter, r *http.Request) {
		err := a.WriteZip(w)
		w.Header().Add("Content-Type", "application/zip")
		if err != nil {
			panic(err)
		}
	})

	hub := room.NewHub(webscene, &movelVersion)
	go hub.Run()

	http.Handle("/live", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWs(w, r)
	}))

	connection := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Serving over: http://%s\n", connection)
	return http.ListenAndServe(connection, nil)
}

func (a App) Generate(outputPath string) error {
	for name, p := range a.Producers {
		fp := path.Join(outputPath, name)

		// Producer names are paths which can contain subfolders, so be sure
		// the subfolders exist before creating the file
		err := os.MkdirAll(filepath.Dir(fp), os.ModeDir)
		if err != nil {
			return err
		}

		// Create the File
		f, err := os.Create(fp)
		if err != nil {
			return err
		}
		defer f.Close()

		// Write data to file
		arifact := p.Data()
		err = arifact.Write(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a App) Run() error {
	if a.Producers == nil || len(a.Producers) == 0 {
		return errors.New("application has not defined any producers")
	}

	os_setup(&a)

	commandMap := make(map[string]*cliCommand)

	var commands []*cliCommand
	commands = []*cliCommand{
		{
			Name:        "Generate",
			Description: "Runs all producers the app has defined and saves it to the file system",
			Aliases:     []string{"generate", "gen"},
			Run: func() error {
				generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
				a.initialize(generateCmd)
				folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
				if err := generateCmd.Parse(os.Args[2:]); err != nil {
					return err
				}
				return a.Generate(*folderFlag)
			},
		},
		{
			Name:        "Serve",
			Description: "Starts an http server and hosts a webplayer for configuring the models produced from this app",
			Aliases:     []string{"serve"},
			Run: func() error {
				serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
				a.initialize(serveCmd)
				portFlag := serveCmd.String("port", "8080", "port to serve over")
				hostFlag := serveCmd.String("host", "localhost", "interface to bind to")

				if err := serveCmd.Parse(os.Args[2:]); err != nil {
					return err
				}
				return a.Serve(*hostFlag, *portFlag)
			},
		},
		{
			Name:        "Outline",
			Description: "Enumerates all generators, parameters, and producers in a heirarchial fashion formatted in JSON",
			Aliases:     []string{"outline"},
			Run: func() error {
				outlineCmd := flag.NewFlagSet("outline", flag.ExitOnError)
				a.initialize(outlineCmd)

				if err := outlineCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				data, err := json.MarshalIndent(a.Schema(), "", "    ")
				if err != nil {
					panic(err)
				}
				os.Stdout.Write(data)

				return nil
			},
		},
		{
			Name:        "Zip",
			Description: "Runs all producers defined and writes it to a zip file",
			Aliases:     []string{"zip", "z"},
			Run: func() error {
				zipCmd := flag.NewFlagSet("zip", flag.ExitOnError)
				a.initialize(zipCmd)
				fileFlag := zipCmd.String("file-name", "out.zip", "file to write the contents of the zip too")

				if err := zipCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				f, err := os.Create(*fileFlag)
				if err != nil {
					return err
				}
				defer f.Close()

				return a.WriteZip(f)
			},
		},
		{
			Name:        "Mermaid",
			Description: "Create a mermaid flow chart for a specific producer",
			Aliases:     []string{"mermaid"},
			Run: func() error {
				mermaidCmd := flag.NewFlagSet("mermaid", flag.ExitOnError)
				a.initialize(mermaidCmd)
				fileFlag := mermaidCmd.String("file-name", "", "Optional path to file to write content to")

				if err := mermaidCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				var out io.Writer = os.Stdout

				if fileFlag != nil && *fileFlag != "" {
					f, err := os.Create(*fileFlag)
					if err != nil {
						return err
					}
					defer f.Close()
					out = f
				}

				return a.WriteMermaid(out)
			},
		},
		{
			Name:        "Help",
			Description: "",
			Aliases:     []string{"help", "h"},
			Run: func() error {
				cliDetails := appCLI{
					Name:        a.Name,
					Version:     a.Version,
					Commands:    commands,
					Authors:     a.Authors,
					Description: a.Description,
				}

				if cliDetails.Version == "" {
					cliDetails.Version = "(no version)"
				}

				tmpl, err := template.New("CLI App").Parse(cliTemplate)
				if err != nil {
					return err
				}
				return tmpl.Execute(os.Stdout, cliDetails)
			},
		},
	}

	for _, cmd := range commands {
		for _, alias := range cmd.Aliases {
			commandMap[alias] = cmd
		}
	}

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) == 0 {
		return commandMap["help"].Run()
	}

	if cmd, ok := commandMap[argsWithoutProg[0]]; ok {
		return cmd.Run()
	}

	fmt.Fprintf(os.Stdout, "unrecognized command %s", argsWithoutProg[0])
	return nil
}
