package generator

import (
	"flag"
	"image"
	"image/png"
	"io"
	"os"
	"path"
	"strings"

	"github.com/EliCDavis/polyform/formats/gltf"
)

type Generator struct {
	Parameters    *GroupParameter
	SubGenerators map[string]*Generator
	Producers     map[string]Producer
}

func (g *Generator) Lookup(path string) *Generator {
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

func (g *Generator) ApplyProfile(profile Profile) ([]*Generator, error) {
	effected := make([]*Generator, 0)
	for genKey, profile := range profile.SubGenerators {
		gen := g.SubGenerators[genKey]
		subsChanged, err := gen.ApplyProfile(profile)
		if err != nil {
			return nil, err
		}
		effected = append(effected, subsChanged...)
	}

	genChanged, err := g.Parameters.ApplyJsonMessage(profile.Parameters)
	if len(effected) > 0 || genChanged {
		effected = append(effected, g)
	}
	return effected, err
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
