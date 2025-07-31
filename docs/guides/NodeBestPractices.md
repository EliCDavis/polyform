Previous: [Manifests](./CreatingManifests.md) | [Table of Contents](../README.md)

# Node Design Best Practices

This guide contains guidelines for creating nodes that will play nice with the rest of the other nodes and graph system.

Have a suggestion on a best practice? Feel free to open up a Issue/PR and start a discussion.

## Keep Things Immutable

A node should never make modifications to both the input provided to it, or the output it produces after it's been returned. Breaking immutability removes the garuntee that a graph's output is reproducable. 

## Avoid Pointers

Avoiding the use of pointers both aids in the effort of keeping things immutable, as well as simplifies the datatypes a node needs to account for.

So instead of this:

```go
type MathNode struct {
	A nodes.Output[*float64]
	B nodes.Output[*float64]
}
```

Prefer:

```go
type MathNode struct {
	A nodes.Output[float64]
	B nodes.Output[float64]
}
```

## Never Panic

Your node outputs should never panic. Doing so halts the execution of the graph, preventing any output from being produced and shown to the user. If you want communicate that something is wrong with the configuration of the node, utilize the `CaptureError` method of the `nodes.StructOutput`, and return something "sensible".

```go
func (mn MathNode) Divide() nodes.StructOutput[float64] {
    a := nodes.TryGetOutputValue(mn.A, 0)
    b := nodes.TryGetOutputValue(mn.B, 0)

    if b == 0 {
        out := nodes.NewStructOutput[float64](0.)
        out.CaptureError(errors.New("can't divide by 0"))
        return out
    }

	return nodes.NewStructOutput(a / b)
}
```

## Make Math Generic When Sensible

There's a lot of basic math operations that can operate over different datatypes. In these scenarios, you can make your nodes generic to take into account of those valid datatypes by using `vector.Number`. This is especially useful for vector math.

So instead of this:

```go
type MathNode struct {
	A nodes.Output[vector3.Float64]
	B nodes.Output[vector3.Float64]
}
```

Prefer

```go
type MathNode[T vector.Number] struct {
	A nodes.Output[vector3.Vector[T]]
	B nodes.Output[vector3.Vector[T]]
}
```

## Prefer float64 and int Data Types

In golang, `int8`, `int16`, `int`, `int32`, `int64`, `float32`, `float64` are all valid datatypes you can do all basic arthimetic with. It'd be messy to create versions of nodes for all of these different datatypes. For that reason, you should opt for using `float64` when you need floating point data, and `int` when you want to restrict your input to whole numbers.

## Use Sensible Defaults For Undefined Inputs

INSERT DESCRIPTION

EXAMPLE: SCALE A VECTOR

```go
func (mn MathNode) Divide() nodes.StructOutput[float64] {
    a := nodes.TryGetOutputValue(mn.A, 0)

    if mn.B == nil {
        return a
    }

    b := mn.B.Value()
    if b == 0 {
        out := nodes.NewStructOutput[float64](0.)
        out.CaptureError(errors.New("can't divide by 0"))
        return out
    }

	return nodes.NewStructOutput(a / b)
}
```

## Give Outputs Meaningful Names

INSERT DESCRIPTION

"opt for sum over out"


## Avoid Calling Input if you can

```go
a := A.Value()
b := B.Value()
if b < 0 { 
    return
}
```

To

```go
b := B.Value()
if b < 0 { 
    return
}
a := A.Value()
```
