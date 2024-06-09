package ply_test

import (
	"bytes"
	"fmt"
	"log"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

var exampleFile = []byte(`ply
format ascii 1.0
comment This is an example comment
obj_info This is example info
element vertex 3
property double x
property double y
property double z
end_header
0 0 0
0 1 0
1 0 0
1 1 0
`)

func ExampleReadHeader() {
	file := bytes.NewBuffer(exampleFile)

	header, _ := ply.ReadHeader(file)

	fmt.Println(header.Format)
	fmt.Println(header.Comments[0])
	fmt.Println(header.ObjInfo[0])

	ele := header.Elements[0]
	fmt.Printf("%s contains %d elements\n", ele.Name, ele.Count)
	fmt.Printf("\t%s", ele.Properties[0].Name())
	fmt.Printf("\t%s", ele.Properties[1].Name())
}

func ExampleReadMesh() {
	file := bytes.NewBuffer(exampleFile)

	mesh, _ := ply.ReadMesh(file)
	log.Println(mesh.Float3Attributes())

	positionData := mesh.Float3Attribute(modeling.PositionAttribute)
	for i := 0; i < positionData.Len(); i++ {
		log.Println(positionData.At(i).Format("%f %f %f"))
	}
}

func ExampleWrite() {
	file := bytes.NewBuffer(exampleFile)
	out := &bytes.Buffer{}

	mesh, _ := ply.ReadMesh(file)
	scaledMesh := mesh.Scale(vector3.New(2., 2., 2.))

	ply.Write(out, scaledMesh, ply.ASCII)
	log.Println(out.String())
}
