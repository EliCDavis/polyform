package nodes

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

func findInFields(fl *ast.FieldList, n ast.Node, tinfo *types.Info, fset *token.FileSet) {
	type FieldReport struct {
		Name string
		Kind string
		Type types.Type
	}
	var reps []FieldReport

	for _, field := range fl.List {
		if field.Names == nil {
			tv, ok := tinfo.Types[field.Type]
			if !ok {
				log.Fatal("not found", field.Type)
			}

			embName := fmt.Sprintf("%v", field.Type)

			_, hostIsStruct := n.(*ast.StructType)
			var kind string

			switch typ := tv.Type.Underlying().(type) {
			case *types.Struct:
				if hostIsStruct {
					kind = "struct (s@s)"
				} else {
					kind = "struct (s@i)"
				}
				reps = append(reps, FieldReport{embName, kind, typ})
			case *types.Interface:
				if hostIsStruct {
					kind = "interface (i@s)"
				} else {
					kind = "interface (i@i)"
				}
				reps = append(reps, FieldReport{embName, kind, typ})
			case *types.Basic:
				if hostIsStruct {
					kind = "basic (b@s)"
				} else {
					kind = "basic (b@i)"
				}
				reps = append(reps, FieldReport{embName, kind, typ})
			case *types.Pointer:
				if hostIsStruct {
					kind = "pointer (p@s)"
				} else {
					kind = "pointer (p@i)"
				}
				reps = append(reps, FieldReport{embName, kind, typ})

			case *types.Array:
				if hostIsStruct {
					kind = "array (a@s)"
				} else {
					kind = "array (a@i)"
				}
				reps = append(reps, FieldReport{embName, kind, typ})
			default:
			}
		}
	}

	if len(reps) > 0 {
		// fmt.Printf("Found at %v\n%v\n", fset.Position(n.Pos()), nodeString(n, fset))

		for _, report := range reps {
			fmt.Printf("--> field '%s' is embedded %s: %s\n", report.Name, report.Kind, report.Type)
		}
		fmt.Println("")
	}
}

func parseFunc() {

}

func findInPackage(pkg *packages.Package, fset *token.FileSet) {
	for _, fileAst := range pkg.Syntax {
		ast.Inspect(fileAst, func(n ast.Node) bool {
			if structTy, ok := n.(*ast.StructType); ok {
				findInFields(structTy.Fields, n, pkg.TypesInfo, fset)
			} else if interfaceTy, ok := n.(*ast.InterfaceType); ok {
				findInFields(interfaceTy.Methods, n, pkg.TypesInfo, fset)
			} else if funcTy, ok := n.(*ast.FuncType); ok {
				log.Printf("Params: %d", len(funcTy.Params.List))
				findInFields(funcTy.Params, n, pkg.TypesInfo, fset)

				log.Printf("Results: %d", len(funcTy.Results.List))
				findInFields(funcTy.Results, n, pkg.TypesInfo, fset)
			}

			return true
		})
	}
}

func Convert(patterns []string) string {
	var tags []string = nil
	var fset = token.NewFileSet()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		// TODO: Need to think about constants in test files. Maybe write type_string_test.go
		// in a separate pass? For later.
		Tests:      false,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
		// Logf:       g.logf,
		Fset: fset,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		log.Fatal(err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages matching %v", len(pkgs), strings.Join(patterns, " "))
	}

	findInPackage(pkgs[0], fset)

	return ""
}
