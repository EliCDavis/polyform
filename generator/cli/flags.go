package cli

import (
	"flag"
	"time"
)

type Flag interface {
	add(*flag.FlagSet)
	value() any
	name() string
	required() bool
	set() bool
	action() error
}

// ============================================================================

type StringFlag struct {
	Name        string
	Value       string
	Description string
	Required    bool
	Action      func(s string) error

	parsedValue *string
}

func (f *StringFlag) add(set *flag.FlagSet) {
	f.parsedValue = set.String(f.Name, f.Value, f.Description)
}

func (f *StringFlag) value() any {
	return *f.parsedValue
}

func (f *StringFlag) name() string {
	return f.Name
}

func (f *StringFlag) required() bool {
	return f.Required
}

func (f *StringFlag) set() bool {
	return f.value() != ""
}

func (f *StringFlag) action() error {
	if f.Action == nil {
		return nil
	}
	return f.Action(*f.parsedValue)
}

// ============================================================================

type BoolFlag struct {
	Name        string
	Value       bool
	Description string
	Action      func(s bool) error

	parsedValue *bool
}

func (f *BoolFlag) add(set *flag.FlagSet) {
	f.parsedValue = set.Bool(f.Name, f.Value, f.Description)
}

func (f *BoolFlag) value() any {
	return *f.parsedValue
}

func (f *BoolFlag) name() string {
	return f.Name
}

func (f *BoolFlag) required() bool {
	return false
}

func (f *BoolFlag) set() bool {
	return true
}

func (f *BoolFlag) action() error {
	if f.Action == nil {
		return nil
	}
	return f.Action(*f.parsedValue)
}

// ============================================================================

type DurationFlag struct {
	Name        string
	Value       time.Duration
	Description string
	Required    bool
	Action      func(s time.Duration) error

	parsedValue *time.Duration
}

func (f *DurationFlag) add(set *flag.FlagSet) {
	f.parsedValue = set.Duration(f.Name, f.Value, f.Description)
}

func (f *DurationFlag) value() any {
	return *f.parsedValue
}

func (f *DurationFlag) name() string {
	return f.Name
}

func (f *DurationFlag) required() bool {
	return f.Required
}

func (f *DurationFlag) set() bool {
	return f.value() != time.Duration(0)
}

func (f *DurationFlag) action() error {
	if f.Action == nil {
		return nil
	}
	return f.Action(*f.parsedValue)
}

// ============================================================================

type Int64Flag struct {
	Name        string
	Value       int64
	Description string
	Required    bool
	Action      func(s int64) error

	parsedValue *int64
}

func (f *Int64Flag) add(set *flag.FlagSet) {
	f.parsedValue = set.Int64(f.Name, f.Value, f.Description)
}

func (f *Int64Flag) value() any {
	return *f.parsedValue
}

func (f *Int64Flag) name() string {
	return f.Name
}

func (f *Int64Flag) required() bool {
	return f.Required
}

func (f *Int64Flag) set() bool {
	return f.value() != int64(0)
}

func (f *Int64Flag) action() error {
	if f.Action == nil {
		return nil
	}
	return f.Action(*f.parsedValue)
}
