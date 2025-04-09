package generator

import (
	"fmt"
	"io"
	"sort"
	"strconv"

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

func (dw DocumentationWriter) writeSingle(writer markdown.Writer) error {
	instances := dw.nodeInstances()
	sections := make(map[string][]schema.NodeType)
	for _, instance := range instances {
		builtSchema := graph.BuildNodeTypeSchema(instance)
		if _, ok := sections[builtSchema.Path]; !ok {
			sections[builtSchema.Path] = make([]schema.NodeType, 0)
		}
		sections[builtSchema.Path] = append(sections[builtSchema.Path], builtSchema)
	}

	nodeCount := 0
	for _, section := range sections {
		nodeCount += len(section)
	}

	sortedSections := utils.SortMapByKey(sections)

	writer.Header1(dw.Title)

	version := dw.Version
	if version == "" {
		version = "(undefined)"
	}

	writer.StartItalics()
	writer.Text(fmt.Sprintf("Version: %s", version))
	writer.EndItalics()
	writer.NewLine()
	writer.NewLine()

	writer.Paragraph(dw.Description)

	writer.Header2("Table Of Contents")

	writer.StartItalics()
	writer.Text(fmt.Sprintf("%d nodes across %d packages", nodeCount, len(sections)))
	writer.EndItalics()
	writer.NewLine()
	writer.NewLine()

	writer.StartBulletList()
	for sectionIndex, section := range sortedSections {

		instances := section.Val
		sort.Slice(instances, func(i, j int) bool {
			return instances[i].DisplayName > instances[j].DisplayName
		})
		sortedSections[sectionIndex].Val = instances

		writer.StartBullet()
		writer.Link(section.Key, strconv.Itoa(sectionIndex))
		writer.EndBullet()

		writer.StartBulletList()
		for instanceIndex, instance := range section.Val {
			writer.StartBullet()
			writer.Link(instance.DisplayName, fmt.Sprintf("%d-%d", sectionIndex, instanceIndex))
			writer.EndBullet()
		}
		writer.EndBulletList()
	}
	writer.EndBulletList()

	for sectionIndex, section := range sortedSections {
		writer.Header2WithId(section.Key, strconv.Itoa(sectionIndex))
		for instanceIndex, instance := range section.Val {
			writer.Header3WithId(instance.DisplayName, fmt.Sprintf("%d-%d", sectionIndex, instanceIndex))

			if instance.Info != "" {
				writer.Paragraph(instance.Info)
			}

			writer.Paragraph("Inputs:")
			if len(instance.Inputs) > 0 {
				writer.StartBulletList()
				sortedInput := utils.SortMapByKey(instance.Inputs)
				for _, input := range sortedInput {
					writer.StartBullet()

					writer.StartBold()
					writer.Text(input.Key)
					writer.EndBold()

					writer.Text(fmt.Sprintf(": %s", input.Val.Type))

					if input.Val.Description != "" {
						writer.Text(" - ")
						writer.Text(input.Val.Description)
					}

					writer.EndBullet()
				}
				writer.EndBulletList()
			} else {
				writer.Paragraph("(None)")
			}

			writer.Paragraph("Outputs:")
			if len(instance.Outputs) > 0 {
				writer.StartBulletList()
				for val, o := range instance.Outputs {
					writer.StartBullet()

					writer.StartBold()
					writer.Text(val)
					writer.EndBold()

					writer.Text(fmt.Sprintf(": %s", o.Type))
					writer.EndBullet()
				}
				writer.EndBulletList()
			} else {
				writer.Paragraph("(None)")
			}
		}
	}

	return writer.Error()
}

func (dw DocumentationWriter) WriteSingleMarkdown(out io.Writer) error {
	writer := markdown.NewWriter(out)
	return dw.writeSingle(writer)
}

func (dw DocumentationWriter) WriteSingleHTML(out io.Writer) error {
	writer := markdown.NewHtmlWriter(out)
	return dw.writeSingle(writer)
}
