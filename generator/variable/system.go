package variable

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/EliCDavis/jbtf"
	"github.com/EliCDavis/polyform/generator/schema"
)

type Profile map[string]json.RawMessage

// System => Info
// Info => Variable
// Variable => Info
// Variable => Reference
// Reference => Variable

// System Kinda copies/inspired by file systems
type System interface {
	Variables() []Info
	Add(path string, variable Variable) error
	Variable(path string) (Variable, error)
	Info(path string) (Info, error)
	Exists(path string) bool
	Remove(path string) error
	Move(oldName, newName string) error
	Traverse(func(path string, info Info, v Variable))
	PersistedSchema(encoder *jbtf.Encoder) (schema.NestedGroup[schema.PersistedVariable], error)
	RuntimeSchema() (schema.NestedGroup[schema.RuntimeVariable], error)
	ApplyProfile(profile Profile) error
	GetProfile() Profile
}

func NewSystem() System {
	return &system{
		entries: make(map[string]systemEntry),
	}
}

type systemEntry interface {
	SetPath(newPath string)
	ApplyProfile(profile json.RawMessage) error
	GetProfile() json.RawMessage
}

type systemVariableEntry struct {
	variable Variable
	info     *info
}

func (sve *systemVariableEntry) SetPath(newPath string) {
	sve.info.name = path.Base(newPath)
}

func (sve *systemVariableEntry) ApplyProfile(profile json.RawMessage) error {
	return sve.variable.applyProfile(profile)
}

func (sve *systemVariableEntry) GetProfile() json.RawMessage {
	return sve.variable.getProfile()
}

type system struct {
	entries map[string]systemEntry
	mutex   sync.RWMutex
}

func (s *system) GetProfile() Profile {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(Profile)

	for key, data := range s.entries {
		result[key] = data.GetProfile()
	}

	return result
}

func (s *system) ApplyProfile(profile Profile) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for key, data := range profile {
		entry, ok := s.entries[key]

		// TODO: Do we really just want to skip over this?
		if !ok {
			continue
		}

		if err := entry.ApplyProfile(data); err != nil {
			return err
		}
	}
	return nil
}

func (s *system) RuntimeSchema() (schema.NestedGroup[schema.RuntimeVariable], error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	variables := make(map[string]schema.RuntimeVariable)
	subgroups := make(map[string]schema.NestedGroup[schema.RuntimeVariable])

	for name, entry := range s.entries {
		switch v := entry.(type) {
		case *systemVariableEntry:
			variables[name] = v.variable.runtimeSchema()

		default:
			panic(fmt.Errorf("unimplemented system entry: %v", entry))
		}
	}

	return schema.NestedGroup[schema.RuntimeVariable]{
		Variables: variables,
		SubGroups: subgroups,
	}, nil
}

func (s *system) PersistedSchema(encoder *jbtf.Encoder) (schema.NestedGroup[schema.PersistedVariable], error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	variables := make(map[string]schema.PersistedVariable)
	subgroups := make(map[string]schema.NestedGroup[schema.PersistedVariable])

	for name, entry := range s.entries {
		switch v := entry.(type) {

		case *systemVariableEntry:
			data, err := v.variable.toPersistantJSON(encoder)
			// data, err := json.Marshal(v.variable)
			if err != nil {
				return schema.NestedGroup[schema.PersistedVariable]{}, err
			}
			variables[name] = schema.PersistedVariable{
				Description: v.info.description,
				Data:        data,
			}

		default:
			panic(fmt.Errorf("unimplemented system entry: %v", entry))
		}
	}

	return schema.NestedGroup[schema.PersistedVariable]{
		Variables: variables,
		SubGroups: subgroups,
	}, nil
}

func (s *system) Variables() []Info {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	out := make([]Info, len(s.entries))
	for _, entry := range s.entries {
		switch v := entry.(type) {

		case *systemVariableEntry:
			out = append(out, v.info)

		default:
			panic(fmt.Errorf("unimplemented system entry: %v", entry))
		}
	}
	return out
}

func (s *system) Move(oldName, newName string) error {
	clean := strings.TrimSpace(newName)
	if clean == "" {
		return fmt.Errorf("new path can not be empty")
	}

	if oldName == clean {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	entry, ok := s.entries[oldName]
	if !ok {
		return fmt.Errorf("variable does not exist at path: %s", oldName)
	}

	if _, ok := s.entries[clean]; ok {
		return fmt.Errorf("variable already exists at path: %s", clean)
	}

	delete(s.entries, oldName)
	entry.SetPath(newName)
	s.entries[clean] = entry
	return nil
}

func (s *system) Add(systemPath string, variable Variable) error {
	if variable == nil {
		return fmt.Errorf("variable is nil")
	}

	clean := strings.TrimSpace(systemPath)
	if clean == "" {
		return fmt.Errorf("can not add variable to empty path")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if _, ok := s.entries[clean]; ok {
		return fmt.Errorf("variable already exists at path: %s", clean)
	}

	info := &info{
		name: path.Base(clean),
	}
	if err := variable.setInfo(info); err != nil {
		return err
	}
	s.entries[clean] = &systemVariableEntry{
		variable: variable,
		info:     info,
	}
	return nil
}

func (s *system) Remove(path string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, ok := s.entries[path]
	if !ok {
		return fmt.Errorf("variable does not exist at path: %s", path)
	}

	delete(s.entries, path)
	return nil
}

func (s *system) Variable(path string) (Variable, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, ok := s.entries[path]
	if !ok {
		return nil, fmt.Errorf("variable does not exist at path: %s", path)
	}

	variableEntry, ok := entry.(*systemVariableEntry)
	if !ok {
		return nil, fmt.Errorf("The path provided %q does not resolve to a variable", path)
	}

	return variableEntry.variable, nil
}

func (s *system) Info(path string) (Info, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, ok := s.entries[path]
	if !ok {
		return nil, fmt.Errorf("variable does not exist at path: %s", path)
	}

	variableEntry, ok := entry.(*systemVariableEntry)
	if !ok {
		return nil, fmt.Errorf("The path provided %q does not resolve to a variable", path)
	}

	return variableEntry.info, nil
}

func (s *system) Exists(path string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	_, ok := s.entries[path]
	return ok
}

func (s *system) Traverse(f func(path string, info Info, v Variable)) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for path, entry := range s.entries {
		switch v := entry.(type) {

		case *systemVariableEntry:
			f(path, v.info, v.variable)

		default:
			panic(fmt.Errorf("unimplemented system entry: %v", entry))
		}
	}
}
