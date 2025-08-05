Previous: [Creating a Custom Node](../CreatingNodes/README.md) | [Table of Contents](../../README.md) | Next: [Best Practices](../NodeBestPractices/README.md)

# Creating Manifest Nodes

Polyform operates on [Directed Acyclic Graphs (DAGs)](https://en.wikipedia.org/wiki/Directed_acyclic_graph), where each node performs a specific task and connects to other nodes through inputs and outputs.

Manifest nodes are special types of nodes that **define the final output of a graph**, typically writing one or more files to the filesystem. 

## What is a Manifest?

A **manifest** is a structure that encapsulates **a set of named entries**, each representing an output artifact such as an image, mesh, or any other serializable content. It also optionally defines a **"main" entry**, which can help tell viewers the first entry to download. 

For example, if we have a manifest for a textured GLTF, you might have 2 entries: the GLTF itself, and a seperate image file the GLTF references. In this scenario, we'd set the GLTF file to be the main entry that a viewer would initially download.

Here's the core structure used for manifests:

```go
type Entry struct {
	Metadata map[string]any `json:"metadata"` // Descriptive metadata about the artifact
	Artifact Artifact       `json:"-"`        // The actual data to be written
}

type Manifest struct {
	Main    string           `json:"main"`    // Optional: name of the main entry
	Entries map[string]Entry `json:"entries"` // All entries in the manifest
}
```

## Artifacts

The `Artifact` interface defines what it means to be a serializable output. Any type that implements this interface can be written as part of a manifest.

```go
type Artifact interface {
	Write(io.Writer) error // Defines how the artifact is written to disk or elsewhere
	Mime() string          // Describes the type of content (e.g., "image/png", "application/json")
}
```

Here's a minimal example for a text file:

```go
type TextArtifact struct {
	Content string
}

func (ta TextArtifact) Write(w io.Writer) error {
	_, err := io.WriteString(w, ta.Content)
	return err
}

func (ta TextArtifact) Mime() string {
	return "text/plain"
}
```

Then you can include this in your manifest entry:

```go
entry := manifest.Entry{
	Metadata: map[string]any{"type": "text"},
	Artifact: TextArtifact{Content: "Hello World!"},
}
```

## Creating a Manifest Node

A manifest node in Polyform is simply a node whose **output type is `manifest.Manifest`**. When Polyform runs a graph and reaches this node, it knows how to gather the manifest entries and process them as final outputs.

Here’s an example of what a simple manifest node might look like:

```go
type MyManifestNode struct {
	Image nodes.Output[image.Image] `description:"The image to export"`
	Path  nodes.Output[string]      `description:"Path to export to"`
}

func (mmn MyManifestNode) Output(out *nodes.StructOutput[manifest.Manifest]) {
	img := nodes.TryGetOutputValue(out, mmn.Image, nil)
	if img == nil {
		return
	}

	entry := manifest.Entry{
		Metadata: map[string]any{
			"description": "Exported PNG image",
		},
		Artifact: &manifest.ImageArtifact{
			Image: img,
			MimeType: "image/png",
		},
	}

	entryPath := nodes.TryGetOutputValue(out, mmn.Path, "export.png")
	out.Set(manifest.Manifest{
		Main: entryPath,
		Entries: map[string]manifest.Entry{
			entryPath: entry,
		},
	})
}
```

> **Note**: `manifest.ImageArtifact` in this example is a hypothetical implementation of the `Artifact` interface that knows how to write an image as PNG.