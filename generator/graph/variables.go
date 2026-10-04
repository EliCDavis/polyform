package graph

import (
	"fmt"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/named"
	"github.com/EliCDavis/polyform/generator/variable"
	"github.com/EliCDavis/polyform/generator/variant"
	"github.com/EliCDavis/polyform/nodes"
)

// NewVariable returns the node type its reference nodes are created with.
func (a *Instance) NewVariable(variablePath string, variable variable.Variable) (string, error) {
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.newVariable(variablePath, variable)
}

func (a *Instance) newVariable(variablePath string, variable variable.Variable) (string, error) {
	if variable == nil {
		return "", fmt.Errorf("trying to add a nil variable %q to graph", variablePath)
	}
	if err := a.variables.Add(variablePath, variable); err != nil {
		return "", fmt.Errorf("failed to add variable to graph: %w", err)
	}

	a.typeFactory.RegisterBuilder(variablePath, func() any {
		return variable.NodeReference()
	})
	nodes.DiscoverNodePortTypes(variable.NodeReference())

	return variablePath, nil
}

// DeleteVariable also deletes every node referencing it.
func (a *Instance) DeleteVariable(variablePath string) error {
	a.mu().Lock()
	defer a.mu().Unlock()

	deleted, err := a.variables.Variable(variablePath)
	if err != nil {
		return fmt.Errorf("trying to delete a variable at the path %q which doesn't contain a variable", variablePath)
	}

	references := make([]nodes.Node, 0)
	for node := range a.nodeIDs {
		ref, ok := node.(variable.Reference)
		if ok && ref.Reference() == deleted {
			references = append(references, node)
		}
	}
	for _, node := range references {
		if _, err := a.deleteNode(node); err != nil {
			return err
		}
	}

	if err := a.variables.Remove(variablePath); err != nil {
		return err
	}
	a.typeFactory.Unregister(variablePath)
	a.incModelVersion()
	return nil
}

func (a *Instance) GetVariable(variablePath string) (variable.Variable, error) {
	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return nil, fmt.Errorf("trying to get a variable at the path %q which doesn't exist", variablePath)
	}
	return variable, nil
}

func (a *Instance) SetVariableInfo(variablePath, newPath, description string) error {
	if err := a.SetVariableDescription(variablePath, description); err != nil {
		return err
	}
	a.mu().Lock()
	defer a.mu().Unlock()
	return a.variables.Move(variablePath, newPath)
}

func (a *Instance) SetVariableDescription(variablePath, description string) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return err
	}
	variable.Info().SetDescription(description)
	return nil
}

func (a *Instance) UpdateVariable(variablePath string, data []byte) (bool, error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return false, err
	}

	r, err := variable.ApplyMessage(data)
	a.incModelVersion()
	return r, err
}

func (a *Instance) VariableData(variablePath string) ([]byte, error) {
	a.mu().Lock()
	defer a.mu().Unlock()

	variable, err := a.variables.Variable(variablePath)
	if err != nil {
		return nil, err
	}
	return variable.ToMessage(), nil
}

func (a *Instance) SwaggerDefinition() swagger.Definition {
	a.mu().RLock()
	defer a.mu().RUnlock()
	return a.variables.SwaggerDefinition()
}

func (a *Instance) Profiles() *named.Collection[variable.Profile] {
	return a.profiles
}

func (a *Instance) SaveProfile(profileName string) {
	a.mu().Lock()
	defer a.mu().Unlock()
	a.Profiles().Set(profileName, a.variables.GetProfile())
}

func (a *Instance) LoadProfile(profileName string) error {
	profile, err := a.Profiles().Get(profileName)
	if err != nil {
		return err
	}
	return a.ApplyProfile(profile)
}

func (a *Instance) ApplyProfile(profile variable.Profile) error {
	a.mu().Lock()
	defer a.mu().Unlock()
	a.incModelVersion()
	return a.variables.ApplyProfile(profile)
}

func (a *Instance) VariantSets() *named.Collection[variant.Set] {
	return a.variantSets
}
