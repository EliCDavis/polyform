package generator

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/EliCDavis/polyform/generator/room"
)

type producerCache map[*Generator]map[string]Artifact

func (pc *producerCache) Lookup(generator *Generator, producer string) Artifact {
	if pc == nil {
		return nil
	}

	var generatorCache map[string]Artifact
	if data, ok := (*pc)[generator]; ok {
		generatorCache = data
	} else {
		return nil
	}

	if artifact, ok := generatorCache[producer]; ok {
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
	Generator   *Generator

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

func writeGeneratorToZip(path string, generator *Generator, zw *zip.Writer, cache producerCache) error {
	if generator == nil {
		panic("can't write nil generator")
	}

	if zw == nil {
		panic("can't write to nil zip writer")
	}

	ctx := &Context{
		Parameters: generator.Parameters,
	}

	for file, producer := range generator.Producers {
		filePath := path + file
		// log.Println(filePath)
		f, err := zw.Create(filePath)
		if err != nil {
			return err
		}

		artifact := cache.Lookup(generator, file)
		if artifact == nil {
			artifact, err = producer(ctx)
			if err != nil {
				return err
			}
		}

		err = artifact.Write(f)
		if err != nil {
			return err
		}
		// log.Printf("wrote %s", filePath)
	}

	for name, gen := range generator.SubGenerators {
		err := writeGeneratorToZip(fmt.Sprintf("%s%s/", path, name), gen, zw, cache)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a App) WriteZip(out io.Writer) error {
	z := zip.NewWriter(out)

	err := writeGeneratorToZip("", a.Generator, z, a.producerCache)
	if err != nil {
		return err
	}

	return z.Close()
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

func (a *App) Serve(host, port string) error {

	a.producerCache = make(map[*Generator]map[string]Artifact)
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
			a.producerCache[g] = make(map[string]Artifact)
		}

		movelVersion++
		w.Write([]byte("{}"))
	})

	http.HandleFunc("/producer/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", "no-cache")

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

		if _, ok := a.producerCache[generatorToUse]; !ok {
			a.producerCache[generatorToUse] = make(map[string]Artifact)
		}

		if artifact, ok := a.producerCache[generatorToUse][producerToLoad]; ok && artifact != nil {
			artifact.Write(w)
			return
		}

		artifact, err := producer(&Context{
			Parameters: generatorToUse.Parameters,
		})

		if err != nil {
			log.Printf(err.Error())
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			err := writeJSONError(w, err)
			if err != nil {
				panic(err)
			}
			return
		}

		err = artifact.Write(w)
		if err != nil {
			panic(err)
		}

		a.producerCache[generatorToUse][producerToLoad] = artifact
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

func (a App) Run() error {
	if a.Generator == nil {
		return errors.New("application has not defined any generators")
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
				a.Generator.initialize(generateCmd)
				folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
				if err := generateCmd.Parse(os.Args[2:]); err != nil {
					return err
				}
				return a.Generator.run(*folderFlag)
			},
		},
		{
			Name:        "Serve",
			Description: "Starts an http server and hosts a webplayer for configuring the models produced from this app",
			Aliases:     []string{"serve"},
			Run: func() error {
				serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
				a.Generator.initialize(serveCmd)
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
				a.Generator.initialize(outlineCmd)

				if err := outlineCmd.Parse(os.Args[2:]); err != nil {
					return err
				}

				data, err := json.MarshalIndent(a.Generator.Schema(), "", "    ")
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
				a.Generator.initialize(zipCmd)
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
