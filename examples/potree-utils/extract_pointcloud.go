package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/urfave/cli/v2"
)

func buildModelWorker(
	ctx *cli.Context,
	nodes <-chan *potree.OctreeNode,
	metadata *potree.Metadata,
	largestPointCount int,
	plyOut io.Writer,
	mutex *sync.Mutex,
	out chan<- int,
) {
	octreeFile, err := openOctreeFile(ctx)
	if err != nil {
		panic(err)
	}
	defer octreeFile.Close()
	pointsProcessed := 0
	potreeBuf := make([]byte, largestPointCount*metadata.BytesPerPoint())
	plyBuf := make([]byte, 27)
	for node := range nodes {
		_, err := octreeFile.Seek(int64(node.ByteOffset), 0)
		if err != nil {
			panic(err)
		}

		_, err = io.ReadFull(octreeFile, potreeBuf[:node.ByteSize])
		if err != nil {
			panic(err)
		}

		cloud := potree.LoadNode(*node, *metadata, potreeBuf[:node.ByteSize])
		count := cloud.PrimitiveCount()
		if count == 0 {
			continue
		}
		positions := cloud.Float3Attribute(modeling.PositionAttribute)
		colors := cloud.Float3Attribute(modeling.ColorAttribute)

		mutex.Lock()

		endien := binary.LittleEndian
		for i := 0; i < count; i++ {
			p := positions.At(i)
			bits := math.Float64bits(p.X())
			endien.PutUint64(plyBuf, bits)
			bits = math.Float64bits(p.Y())
			endien.PutUint64(plyBuf[8:], bits)
			bits = math.Float64bits(p.Z())
			endien.PutUint64(plyBuf[16:], bits)

			c := colors.At(i).Scale(255)
			plyBuf[24] = byte(c.X())
			plyBuf[25] = byte(c.Y())
			plyBuf[26] = byte(c.Z())

			_, err = plyOut.Write(plyBuf)
			if err != nil {
				panic(err)
			}
		}

		mutex.Unlock()

		pointsProcessed += count
	}

	out <- pointsProcessed
}

func buildModelWithChildren(ctx *cli.Context, root *potree.OctreeNode, metadata *potree.Metadata) error {

	f, err := os.Create(ctx.String("out"))
	if err != nil {
		return err
	}
	defer f.Close()

	largestPointCount := 0
	root.Walk(func(o *potree.OctreeNode) {
		if o.NumPoints > uint32(largestPointCount) {
			largestPointCount = int(o.NumPoints)
		}
	})

	header := ply.Header{
		Format: ply.BinaryLittleEndian,
		Elements: []ply.Element{
			{
				Name:  ply.VertexElementName,
				Count: int64(root.PointCount()),
				Properties: []ply.Property{
					&ply.ScalarProperty{PropertyName: "x", Type: ply.Double},
					&ply.ScalarProperty{PropertyName: "y", Type: ply.Double},
					&ply.ScalarProperty{PropertyName: "z", Type: ply.Double},
					&ply.ScalarProperty{PropertyName: "red", Type: ply.UChar},
					&ply.ScalarProperty{PropertyName: "green", Type: ply.UChar},
					&ply.ScalarProperty{PropertyName: "blue", Type: ply.UChar},
				},
			},
		},
	}

	bufWriter := bufio.NewWriter(f)
	err = header.Write(bufWriter)
	if err != nil {
		return err
	}

	workerCount := runtime.NumCPU()
	jobs := make(chan *potree.OctreeNode, workerCount)
	meshes := make(chan int, workerCount)

	mutex := &sync.Mutex{}
	for i := 0; i < workerCount; i++ {
		go buildModelWorker(ctx, jobs, metadata, largestPointCount, bufWriter, mutex, meshes)
	}

	root.Walk(func(o *potree.OctreeNode) { jobs <- o })
	close(jobs)

	for i := 0; i < workerCount; i++ {
		<-meshes
	}

	return bufWriter.Flush()
}

func buildModel(ctx *cli.Context, node *potree.OctreeNode, metadata *potree.Metadata) (*modeling.Mesh, error) {
	octreeFile, err := openOctreeFile(ctx)
	if err != nil {
		return nil, err
	}
	defer octreeFile.Close()

	_, err = octreeFile.Seek(int64(node.ByteOffset), 0)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, node.ByteSize)
	_, err = io.ReadFull(octreeFile, buf)
	if err != nil {
		return nil, err
	}

	mesh := potree.LoadNode(*node, *metadata, buf)
	return &mesh, nil
}

func findNode(name string, node *potree.OctreeNode) (*potree.OctreeNode, error) {
	if name == node.Name {
		return node, nil
	}

	for _, c := range node.Children {
		if strings.Index(name, c.Name) == 0 {
			return findNode(name, c)
		}
		log.Println(c.Name)
	}

	return nil, fmt.Errorf("%s can't find child node with name %s", node.Name, name)
}

var ExtractPointcloudCommand = &cli.Command{
	Name: "extract-pointcloud",
	Flags: []cli.Flag{
		metadataFlag,
		hierarchyFlag,
		octreeFlag,
		&cli.StringFlag{
			Name:  "node",
			Value: "r",
			Usage: "Name of node to extract point data from",
		},
		&cli.BoolFlag{
			Name:  "include-children",
			Value: false,
			Usage: "Whether or not to include children data",
		},
		&cli.StringFlag{
			Name:  "out",
			Value: "out.ply",
			Usage: "Name of ply file to write pointcloud data too",
		},
	},
	Action: func(ctx *cli.Context) error {
		metadata, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}

		startNode, err := findNode(ctx.String("node"), hierarchy)
		if err != nil {
			return err
		}

		if ctx.Bool("include-children") {
			start := time.Now()
			err = buildModelWithChildren(ctx, startNode, metadata)
			log.Printf("PLY written in %s", time.Since(start))
			return err
		}

		mesh, err := buildModel(ctx, startNode, metadata)
		if err != nil {
			return err
		}
		fmt.Fprintf(ctx.App.Writer, "Writing pointcloud with %d points to %s", mesh.Indices().Len(), ctx.String("out"))
		return ply.SaveBinary(ctx.String("out"), *mesh)
	},
}
