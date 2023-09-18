package generator

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	_ "embed"

	"github.com/EliCDavis/polyform/formats/gltf"
)

type Generator struct {
	Parameters    *GroupParameter
	SubGenerators map[string]Generator
	Producers     map[string]Producer
}

func (g Generator) Lookup(path string) Generator {
	if path == "" {
		return g
	}

	startSplit := strings.Index(path, "/")

	if startSplit == -1 {
		return g.SubGenerators[path]
	}

	return g.SubGenerators[path[:startSplit]].Lookup(path[startSplit+1:])

}

func (g Generator) Schema() GeneratorSchema {
	schema := GeneratorSchema{
		Producers:     make([]string, 0, len(g.Producers)),
		SubGenerators: make(map[string]GeneratorSchema),
	}

	if g.Parameters != nil {
		schema.Parameters = g.Parameters.GroupParameterSchema()
	}

	for key := range g.Producers {
		schema.Producers = append(schema.Producers, key)
	}

	for key, val := range g.SubGenerators {
		schema.SubGenerators[key] = val.Schema()
	}

	return schema
}

func (g Generator) initialize(set *flag.FlagSet) {
	for _, g := range g.SubGenerators {
		g.initialize(set)
	}

	if g.Parameters != nil {
		g.Parameters.initializeForCLI(set)
	}
}

func (g Generator) run(outputPath string) error {

	// Run Sub Generators First
	for key, subG := range g.SubGenerators {
		err := subG.run(path.Join(outputPath, key))
		if err != nil {
			return err
		}
	}

	// Initialize Context
	ctx := &Context{
		Parameters: g.Parameters,
	}

	err := os.MkdirAll(outputPath, os.ModeDir)
	if err != nil {
		return err
	}

	// Run Producers
	for name, p := range g.Producers {
		arifact, err := p(ctx)
		if err != nil {
			return err
		}
		f, err := os.Create(path.Join(outputPath, name))
		if err != nil {
			return err
		}
		defer f.Close()
		err = arifact.Write(f)
		if err != nil {
			return err
		}
	}
	return nil
}

//go:embed server.html
var indexData []byte

func (g Generator) ApplyProfile(profile Profile) error {
	for gen, profile := range profile.SubGenerators {
		err := g.SubGenerators[gen].ApplyProfile(profile)
		if err != nil {
			return err
		}
	}

	return g.Parameters.ApplyJsonMessage(profile.Parameters)
}

func (g Generator) Serve(port string) error {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// d, err := os.ReadFile("generator/server.html")
		// if err != nil {
		// 	panic(err)
		// }
		// w.Write(d)
		w.Write(indexData)
	})

	http.HandleFunc("/schema", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(g.Schema())
		if err != nil {
			panic(err)
		}
		w.Write(data)
	})

	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		profile := Profile{}
		if err := json.Unmarshal(body, &profile); err != nil {
			panic(err)
		}
		err := g.ApplyProfile(profile)
		if err != nil {
			panic(err)
		}
		w.Write([]byte("{}"))
	})

	http.HandleFunc("/producer/", func(w http.ResponseWriter, r *http.Request) {
		// params, _ := url.ParseQuery(r.URL.RawQuery)

		generatorToUse := g
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
			panic(fmt.Errorf("No producer registered for: %s", producerToLoad))
		}
		artifact, err := producer(&Context{
			Parameters: generatorToUse.Parameters,
		})
		if err != nil {
			panic(err)
		}
		artifact.Write(w)
	})

	fmt.Printf("Serving over: http://localhost:%s\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func (g Generator) Run() error {

	argsWithoutProg := os.Args[1:]

	switch strings.ToLower(argsWithoutProg[0]) {
	case "generate":
		generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
		g.initialize(generateCmd)
		folderFlag := generateCmd.String("folder", ".", "folder to save generated contents to")
		if err := generateCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return g.run(*folderFlag)

	case "serve":
		serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		g.initialize(serveCmd)
		portFlag := serveCmd.String("port", "8080", "port to serve over")
		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			return err
		}
		return g.Serve(*portFlag)

	default:
		fmt.Fprintf(os.Stdout, "unrecognized command %s", argsWithoutProg[0])
	}

	return nil
}

type Context struct {
	Parameters *GroupParameter
}

type Producer func(c *Context) (Artifact, error)

type Artifact interface {
	Write(io.Writer) error
}

type PolyformArtifact[T any] interface {
	Artifact
	Value() T
}

type ImageArtifact struct {
	Image image.Image
}

func (ia ImageArtifact) Write(w io.Writer) error {
	return png.Encode(w, ia.Image)
}

type GltfArtifact struct {
	Scene gltf.PolyformScene
}

func (ga GltfArtifact) Write(w io.Writer) error {
	return gltf.WriteBinary(ga.Scene, w)
}
