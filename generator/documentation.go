package generator

import (
	"fmt"
	"io"

	"github.com/EliCDavis/polyform/formats/markdown"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/polyform/utils"
)

type DocumentationWriter struct {
	Title       string
	Description string
	Version     string
	NodeTypes   *refutil.TypeFactory
}

func (dw DocumentationWriter) nodeInstances() []nodes.Node {
	registeredTypes := dw.NodeTypes.Types()
	registeredNodes := make([]nodes.Node, 0, len(registeredTypes))

	for _, registeredType := range registeredTypes {
		nodeInstance, ok := dw.NodeTypes.New(registeredType).(nodes.Node)
		if !ok {
			panic(fmt.Errorf("Registered type %q is not a node", registeredType))
		}

		if nodeInstance == nil {
			panic(fmt.Errorf("New registered type %q is nil", registeredType))
		}

		registeredNodes = append(registeredNodes, nodeInstance)
	}

	return registeredNodes
}

func (dw DocumentationWriter) WriteSingleMarkdown(out io.Writer) error {

	instances := dw.nodeInstances()
	sections := make(map[string][]schema.NodeType)
	for _, instance := range instances {
		builtSchema := graph.BuildNodeTypeSchema(instance)
		if _, ok := sections[builtSchema.Path]; !ok {
			sections[builtSchema.Path] = make([]schema.NodeType, 0)
		}
		sections[builtSchema.Path] = append(sections[builtSchema.Path], builtSchema)
	}

	sortedSections := utils.SortMapByKey(sections)

	writer := markdown.NewWriter(out)

	writer.Header1(dw.Title)

	version := dw.Version
	if version == "" {
		version = "(undefined)"
	}
	writer.Paragraph(fmt.Sprintf("*Version: %s*", version))
	writer.Paragraph(dw.Description)

	for _, section := range sortedSections {
		writer.Header2(section.Key)
		for _, instance := range section.Val {
			writer.Header3(instance.DisplayName)

			if instance.Info != "" {
				writer.Paragraph(instance.Info)
			}

			writer.Paragraph("Inputs:")
			if len(instance.Inputs) > 0 {
				for val, i := range instance.Inputs {
					writer.Bullet(fmt.Sprintf("**%s**: %s", val, i.Type))
				}
				writer.NewLine()
			} else {
				writer.Paragraph("(None)")
			}

			writer.Paragraph("Outputs:")
			if len(instance.Outputs) > 0 {
				for val, o := range instance.Outputs {
					writer.Bullet(fmt.Sprintf("**%s**: %s", val, o.Type))
				}
				writer.NewLine()
			} else {
				writer.Paragraph("(None)")
			}
		}
	}

	return writer.Error()
}
