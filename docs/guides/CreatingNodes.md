[Table of Contents](../README.md) | Next: [Manifests](./CreatingManifests.md)

# Creating a Custom Node

In Polyform, a node generally represents a single operation in a procedural graph. 

Inputs to nodes are other nodes' outputs, and outputs are computed dynamically as the graph executes.

## Defining a Node

Nodes generally start as a struct, where the fields of that struct represent it's connections to other node's output.

For the sake of presenting information to the user, as well as automatically generating documentation, you can define descriptions for the node as well as each of it's inputs.

```go
type MathNode struct {
	A nodes.Output[float64] `description:"The first variable"`
	B nodes.Output[float64] `description:"The second variable"`
}

func (MathNode) Description() string {
    return "Performs common math operations that require two variables"
}
```

### Defining Outputs 

Any method on a struct that accepts a `*nodes.StructOutput[T]` is automatically treated as an output port by the system. `StructOutput[T]` acts as a wrapper around the value you want to output for supporting things like error handling.

```go
func (cn MathNode) Add(out *nodes.StructOutput[float64]) {
    a := cn.A.Value()
    b := cn.B.Value()
	out.Set(a + b)
}
```

However, when implementing the function, a node should never assume any of its input ports are set. Calling `Value()` on a `nil` input port will cause the runtime to panic and the graph execution to halt. The utility `nodes.TryGetOutputValue` has been introduced that will attempt to take the value of an input port if it exists, and returns a fallback value when the input port is `nil`. `TryGetOutputValue` also keeps up with how much time it takes for the input to execute, subtracting it from the `MathNode`'s execution time, allowing for proper reporting of performance on a per node basis. Updating our code to be more safe results in our `Add` function looking like:

```go
func (mn MathNode) Add(out *nodes.StructOutput[float64]) {
    a := nodes.TryGetOutputValue(out, mn.A, 0)
    b := nodes.TryGetOutputValue(out, mn.B, 0)
	out.Set(a + b)
}
```

Sometimes, the current input into a node is in effect "invalid" and no real computation can be done. In these scenarios a _sensible_ default value needs to be returned. Unfortunately, what "_sensible_" means is dependent upon the kind of operations being performed by the node, but a good rule of thumb is returning the ["zero" value](https://go.dev/ref/spec#The_zero_value) of the datatype.

Before we return our sensible value, we can capture an error to alert the graph system that something has gone wrong.

```go
func (mn MathNode) Divide(out *nodes.StructOutput[float64]) {
    a := nodes.TryGetOutputValue(out, mn.A, 0)
    b := nodes.TryGetOutputValue(out, mn.B, 0)

    if b == 0 {
		// By default, the output is the zero value already, so this line 
		// effectively acts as a no-op, and is kept for demonstration purposes
		// only.
		out.Set(0) 
        out.CaptureError(errors.New("can't divide by 0"))
        return
    }

	out.Set(a / b)
}
```

To add descriptions to the output ports, define a method on your node struct named `<OutputMethodName>Description() string`. This will attach a tooltip or label to that output in the editor and documentation.

```go
// Describes the Add output port
func (MathNode) AddDescription() string {
    return "Adds A and B together"
}

// Describes the Divide output port
func (MathNode) DivideDescription() string {
    return "Divides A by B, returning 0 if B is undefined or 0"
}
```

### Array Inputs

If the operation you're performing can take any number of inputs, you can define your input as type `[]nodes.Output[T]`. Doing so allows users to wire up multiple nodes into the same input slot. You can then use `nodes.GetOutputValues` to call resolve all inputs, creating timings while doing so.

```go
type SumNode struct {
	Values []nodes.Output[float64] `description:"The nodes to sum"`
}

func (sn SumNode) Sum(out *nodes.StructOutput[float64]) {
	var total float64
	values := nodes.GetOutputValues(out, sn.Values)
	for _, v := range values {
		total += v
	}
	out.Set(total)
}
```

## Registering a Node

Now that you've defined your node, you need to take steps to include it in a build of the graph system.

### Registering the Package

The standard way to register your nodes with the graph system is to define a `init` function in your package. You can read more about it's [specifics of execution here](https://go.dev/doc/effective_go#init).

Inside the init function, you create a `TypeFactory` which collects all the types your package wants to register. Then you pass that factory to `generator.RegisterTypes(factory)` to integrate your custom nodes into Polyform’s node registry.

In this guide, we've been defining a node that takes advantage of the `nodes.Struct[T]` functionality. To have our node work, we need to wrap our type while registering it, resulting in `refutil.RegisterType[nodes.Struct[MathNode]](factory)`.

```go
package mycoolpackage

import (
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[MathNode]](factory)

	generator.RegisterTypes(factory)
}
```

### Including a Package in a Build

For the init function to be called, we need to include the package in a build. The easiest way is to ["import for side effect"](https://go.dev/doc/effective_go#blank_import) by prefixing the package name with a underscore.

You can take a look at the [main entrypoint into polyform](../../cmd/polyform/main.go) for a more elaborate example.

```go
package main

import (
	"fmt"
	"os"

	"github.com/EliCDavis/polyform/generator"

	// Import so they register their nodes with the generator
	_ "mycoolpackage"
)

func main() {
	app := generator.App{ Name: "Custom Polyform" }

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
```