package variable

import (
	"encoding/json"
	"fmt"
	"maps"
	"path"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/formats/swagger"
	"github.com/EliCDavis/polyform/generator/persistence"
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
	PersistedSchema(encoder Encoder) (schema.NestedGroup[persistence.Variable], error)
	RuntimeSchema() (schema.NestedGroup[schema.Variable], error)
	ApplyProfile(profile Profile) error
	Matches(profile Profile) bool
	GetProfile() Profile
	VariableProfile(path string) (json.RawMessage, error)
	PersistProfile(encoder Encoder, profile Profile) (Profile, error)
	LoadProfile(decoder Decoder, profile Profile) (Profile, error)
	PruneBlobs(profiles map[string]Profile)
	SwaggerDefinition() swagger.Definition
}

// Implemented by variables whose profile entries name shared byte content.
type blobProfiled interface {
	persistProfile(encoder Encoder, profile json.RawMessage) (json.RawMessage, error)
	loadProfile(decoder Decoder, profile json.RawMessage) (json.RawMessage, error)
	pruneBlobs(referenced []json.RawMessage)
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
	SwaggerProperty() swagger.Property
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

func (sve *systemVariableEntry) SwaggerProperty() swagger.Property {
	return sve.variable.SwaggerProperty()
}

type system struct {
	entries map[string]systemEntry
	mutex   sync.RWMutex
}

func (s *system) SwaggerDefinition() swagger.Definition {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	props := make(map[string]swagger.Property, len(s.entries))
	for name, entry := range s.entries {
		props[name] = entry.SwaggerProperty()
	}

	return swagger.Definition{
		Type:       "object",
		Properties: props,
	}
}

func (s *system) GetProfile() Profile {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	result := make(Profile)

	for key, data := range s.entries {
		result[key] = data.GetProfile()
	}

	return result
}

func (s *system) VariableProfile(path string) (json.RawMessage, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	entry, ok := s.entries[path]
	if !ok {
		return nil, fmt.Errorf("variable does not exist at path: %s", path)
	}
	return entry.GetProfile(), nil
}

func (s *system) blobVariable(path string) (blobProfiled, bool) {
	entry, ok := s.entries[path].(*systemVariableEntry)
	if !ok {
		return nil, false
	}
	blobs, ok := entry.variable.(blobProfiled)
	return blobs, ok
}

func (s *system) PersistProfile(encoder Encoder, profile Profile) (Profile, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	persisted := make(Profile, len(profile))
	for _, path := range slices.Sorted(maps.Keys(profile)) {
		raw := profile[path]
		if blobs, ok := s.blobVariable(path); ok {
			var err error
			if raw, err = blobs.persistProfile(encoder, raw); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
		}
		persisted[path] = raw
	}
	return persisted, nil
}

func (s *system) LoadProfile(decoder Decoder, profile Profile) (Profile, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	loaded := make(Profile, len(profile))
	for path, raw := range profile {
		if blobs, ok := s.blobVariable(path); ok {
			var err error
			if raw, err = blobs.loadProfile(decoder, raw); err != nil {
				return nil, fmt.Errorf("%s: %w", path, err)
			}
		}
		loaded[path] = raw
	}
	return loaded, nil
}

func (s *system) PruneBlobs(profiles map[string]Profile) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for path := range s.entries {
		blobs, ok := s.blobVariable(path)
		if !ok {
			continue
		}
		referenced := make([]json.RawMessage, 0, len(profiles))
		for _, profile := range profiles {
			if raw, ok := profile[path]; ok {
				referenced = append(referenced, raw)
			}
		}
		blobs.pruneBlobs(referenced)
	}
}

func (s *system) ApplyProfile(profile Profile) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for key, data := range profile {
		entry, ok := s.entries[key]

		if !ok {
			return fmt.Errorf("unable to apply %q from profile, variable does not exist", key)
		}

		if sameJSON(entry.GetProfile(), data) {
			continue
		}

		if err := entry.ApplyProfile(data); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
	}
	return nil
}

func (s *system) Matches(profile Profile) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for key, data := range profile {
		entry, ok := s.entries[key]
		if !ok || !sameJSON(entry.GetProfile(), data) {
			return false
		}
	}
	return true
}

func sameJSON(a, b json.RawMessage) bool {
	var left, right any
	if json.Unmarshal(a, &left) != nil || json.Unmarshal(b, &right) != nil {
		return false
	}
	return reflect.DeepEqual(left, right)
}

func (s *system) RuntimeSchema() (schema.NestedGroup[schema.Variable], error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	variables := make(map[string]schema.Variable)

	for name, entry := range s.entries {
		switch v := entry.(type) {
		case *systemVariableEntry:
			variables[name] = v.variable.runtimeSchema()

		default:
			panic(fmt.Errorf("unimplemented system entry: %v", entry))
		}
	}

	return schema.Nest(variables), nil
}

// Sorted, so equal states produce equal buffer-view numbering.
func (s *system) PersistedSchema(encoder Encoder) (schema.NestedGroup[persistence.Variable], error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	variables := make(map[string]persistence.Variable)

	for _, name := range slices.Sorted(maps.Keys(s.entries)) {
		entry := s.entries[name]
		switch v := entry.(type) {

		case *systemVariableEntry:
			data, err := v.variable.toPersistantJSON(encoder)
			if err != nil {
				return schema.NestedGroup[persistence.Variable]{}, err
			}
			variables[name] = persistence.Variable{
				Description: v.info.description,
				Data:        data,
			}

		default:
			panic(fmt.Errorf("unimplemented system entry: %v", entry))
		}
	}

	return schema.Nest(variables), nil
}

func (s *system) Variables() []Info {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	out := make([]Info, 0, len(s.entries))
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
