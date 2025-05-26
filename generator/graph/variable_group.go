package graph

import (
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/generator/variable"
)

func variablePathComponents(variablePath string) (groupName, subPath string) {
	i := strings.Index(variablePath, "/")

	// No subpathing. No subgroup
	if i == -1 {
		groupName = ""
		subPath = variablePath
		return
	}

	groupName = variablePath[:i]
	subPath = variablePath[i+1:]
	return
}

func NewVariableGroup() *VariableGroup {
	return &VariableGroup{
		variables: make(map[string]variable.Variable),
		subGroups: make(map[string]*VariableGroup),
		mutex:     &sync.RWMutex{},
	}
}

type VariableGroup struct {
	variables map[string]variable.Variable
	subGroups map[string]*VariableGroup

	// TODO: Two mutexes, one per map
	mutex *sync.RWMutex
}

func (vg *VariableGroup) Schema() schema.VariableGroup {
	vg.mutex.RLock()
	defer vg.mutex.RUnlock()

	variableSchema := make(map[string]variable.JsonContainer)
	for name, vars := range vg.variables {
		variableSchema[name] = variable.JsonContainer{
			Variable: vars,
		}
	}

	groupSchema := make(map[string]schema.VariableGroup)
	for name, group := range vg.subGroups {
		groupSchema[name] = group.Schema()
	}

	return schema.VariableGroup{
		Variables: variableSchema,
		SubGroups: groupSchema,
	}
}

func (vg *VariableGroup) AddVariable(variablePath string, variable variable.Variable) {
	vg.mutex.Lock()
	defer vg.mutex.Unlock()

	groupName, subPathName := variablePathComponents(variablePath)

	// No subpathing. Just add to this group
	if groupName == "" {
		vg.variables[subPathName] = variable
		return
	}

	group, ok := vg.subGroups[groupName]
	if ok {
		group.AddVariable(subPathName, variable)
		return
	}

	group = NewVariableGroup()
	group.AddVariable(subPathName, variable)
	vg.subGroups[groupName] = group
}

func (vg *VariableGroup) RemoveVariable(variablePath string) {
	vg.mutex.Lock()
	defer vg.mutex.Unlock()

	groupName, subPathName := variablePathComponents(variablePath)

	if groupName == "" {
		delete(vg.variables, variablePath)
		return
	}

	vg.subGroups[groupName].RemoveVariable(subPathName)
}

func (vg *VariableGroup) HasVariable(variablePath string) bool {
	vg.mutex.RLock()
	defer vg.mutex.RUnlock()

	groupName, subPathName := variablePathComponents(variablePath)

	if groupName == "" {
		_, ok := vg.variables[variablePath]
		return ok
	}

	_, ok := vg.subGroups[groupName]
	if !ok {
		return false
	}
	return vg.subGroups[groupName].HasVariable(subPathName)
}

func (vg *VariableGroup) GetVariable(variablePath string) variable.Variable {
	vg.mutex.RLock()
	defer vg.mutex.RUnlock()

	groupName, subPathName := variablePathComponents(variablePath)

	if groupName == "" {
		return vg.variables[variablePath]
	}

	return vg.subGroups[groupName].GetVariable(subPathName)
}

func (vg *VariableGroup) HasSubgroup(subgroupPath string) bool {
	vg.mutex.RLock()
	defer vg.mutex.RUnlock()

	groupName, subPathName := variablePathComponents(subgroupPath)

	if groupName == "" {
		_, ok := vg.subGroups[subgroupPath]
		return ok
	}

	if _, ok := vg.subGroups[groupName]; !ok {
		return false
	}

	return vg.subGroups[groupName].HasSubgroup(subPathName)
}

func (vg *VariableGroup) GetSubgroup(subgroupPath string) *VariableGroup {
	vg.mutex.RLock()
	defer vg.mutex.RUnlock()

	groupName, subPathName := variablePathComponents(subgroupPath)

	if groupName == "" {
		return vg.subGroups[subgroupPath]
	}

	return vg.subGroups[groupName].GetSubgroup(subPathName)
}

func (vg *VariableGroup) AddSubgroup(subgroupPath string) {
	vg.mutex.Lock()
	defer vg.mutex.Unlock()

	groupName, subPath := variablePathComponents(subgroupPath)

	if groupName == "" {
		vg.subGroups[subgroupPath] = NewVariableGroup()
		return
	}

	g := NewVariableGroup()
	vg.subGroups[groupName] = g
	g.AddSubgroup(subPath)
}

func (vg *VariableGroup) RemoveSubgroup(subgroupPath string) {
	vg.mutex.Lock()
	defer vg.mutex.Unlock()

	groupName, subPath := variablePathComponents(subgroupPath)

	if groupName == "" {
		delete(vg.subGroups, subgroupPath)
		return
	}

	vg.subGroups[groupName].RemoveSubgroup(subPath)
}
