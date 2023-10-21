package generator

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/room"
)

type App struct {
	Name        string
	Version     string
	Description string
	WebScene    *room.WebScene
	Authors     []Author
	Generator   *Generator
}

type pageData struct {
	Title       string
	Version     string
	Description string
	Scripting   string
}

//go:embed html/*
var htmlFs embed.FS

func (a App) Serve(host, port string) error {

	producerCache := make(map[*Generator]map[string]Artifact)
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
		data, err := json.Marshal(a.Generator.Schema())
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
		generatorsEffected, err := a.Generator.ApplyProfile(profile)
		if err != nil {
			panic(err)
		}

		for _, g := range generatorsEffected {
			producerCache[g] = make(map[string]Artifact)
		}

		movelVersion++
		w.Write([]byte("{}"))
	})

	http.HandleFunc("/producer/", func(w http.ResponseWriter, r *http.Request) {
		producerLock.Lock()
		defer producerLock.Unlock()
		// params, _ := url.ParseQuery(r.URL.RawQuery)

		generatorToUse := a.Generator
		components := strings.Split(r.URL.Path, "/")
		for i := 2; i < len(components)-1; i++ {
			newGen, ok := generatorToUse.SubGenerators[components[i]]
			if !ok {
				panic(fmt.Errorf("no sub generator exists: %s", components[i]))
			}
			generatorToUse = newGen
		}

		producerToLoad := path.Base(r.URL.Path)

		producer, ok := generatorToUse.Producers[producerToLoad]
		if !ok {
			panic(fmt.Errorf("no producer registered for: %s", producerToLoad))
		}

		if _, ok := producerCache[generatorToUse]; !ok {
			producerCache[generatorToUse] = make(map[string]Artifact)
		}

		if artifact, ok := producerCache[generatorToUse][producerToLoad]; ok && artifact != nil {
			artifact.Write(w)
			return
		}

		artifact, err := producer(&Context{
			Parameters: generatorToUse.Parameters,
		})
		if err != nil {
			panic(err)
		}
		err = artifact.Write(w)
		if err != nil {
			panic(err)
		}

		producerCache[generatorToUse][producerToLoad] = artifact
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

func (a App) Run() error {

	argsWithoutProg := os.Args[1:]

	switch strings.ToLower(argsWithoutProg[0]) {
	case "generate":
		generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
		a.Generator.initialize(generateCmd)
		folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
		if err := generateCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return a.Generator.run(*folderFlag)

	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		a.Generator.initialize(serveCmd)
		portFlag := serveCmd.String("port", "8080", "port to serve over")
		hostFlag := serveCmd.String("host", "localhost", "interface to bind to")

		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return a.Serve(*hostFlag, *portFlag)

	default:
		fmt.Fprintf(os.Stdout, "unrecognized command %s", argsWithoutProg[0])
	}

	return nil
}
