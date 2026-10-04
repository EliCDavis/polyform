package graph

import (
	gsync "sync"

	"github.com/EliCDavis/polyform/generator/graph/history"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/named"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/EliCDavis/polyform/generator/subgraph"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type Details struct {
	Name        string
	Version     string
	Description string
	Authors     []persistence.Author
}

// An Instance is a whole project: the root graph, the subgraph definitions
// it can place, and everything saved alongside them.
type Instance struct {
	// The root graph.
	*Graph

	details         Details
	typeFactory     *refutil.TypeFactory
	variableFactory func(string) (variable.Variable, error)
	modelVersion    uint32

	variables      variable.System
	profiles       *named.Collection[variable.Profile]
	variantSets    *named.Collection[variant.Set]
	namedManifests *namedOutputManager[manifest.Manifest]
	subGraphs      map[string]*subGraphRuntime
	history        *history.History

	compound    compoundEdit
	copySources map[string]savedGraph
	lock        gsync.RWMutex
}

type Config struct {
	Name            string
	Version         string
	Description     string
	Authors         []persistence.Author
	TypeFactory     *refutil.TypeFactory
	VariableFactory func(string) (variable.Variable, error)
}

func New(config Config) *Instance {
	// Cloned, not shared: a subgraph's runtime builder captures this
	// instance, and callers hand over a registry they reuse.
	factory := &refutil.TypeFactory{}
	if config.TypeFactory != nil {
		factory = factory.Combine(config.TypeFactory)
	}

	nodes.DiscoverPortTypes(factory)
	instance := &Instance{
		typeFactory:     factory,
		variableFactory: config.VariableFactory,
		variables:       variable.NewSystem(),
	}
	instance.Graph = newGraph(instance)
	instance.Reset()
	instance.details = Details{
		Name:        config.Name,
		Description: config.Description,
		Version:     config.Version,
		Authors:     config.Authors,
	}
	instance.history = history.New(instance.EncodeToAppSchema, instance.ApplyAppSchema)
	return instance
}

func (a *Instance) Reset() {
	a.details = Details{
		Name:    "New Graph",
		Version: "v0.0.0",
		Authors: []persistence.Author{},
	}

	a.Graph.clear()

	a.variables.Traverse(func(path string, info variable.Info, v variable.Variable) {
		a.typeFactory.Unregister(path)
	})
	a.variables = variable.NewSystem()
	a.namedManifests = newNamedOutputManager[manifest.Manifest]()
	a.profiles = named.New[variable.Profile]("profile")
	a.variantSets = named.New[variant.Set]("variant set")
	a.subGraphs = make(map[string]*subGraphRuntime)
	a.copySources = make(map[string]savedGraph)

	// A snapshot is restored into a registry holding only what it defines.
	for _, key := range a.typeFactory.Types() {
		if subgraph.IsRuntimeNodeType(key) {
			a.typeFactory.Unregister(key)
		}
	}
}

func (a *Instance) History() *history.History {
	return a.history
}

func (a *Instance) Details() Details {
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.details
}

func (a *Instance) SetDetails(details Details) {
	a.mu().Lock()
	defer a.mu().Unlock()
	a.details = details
}

func (a *Instance) ModelVersion() uint32 {
	return a.modelVersion
}
