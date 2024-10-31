//go:build ignore

package main

import (
	"bytes"
	"cmp"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"
)

// Generates clone.go

type NodeType int

const (
	NodeTypeStruct NodeType = iota
	NodeTypeSlice
)

type CloneableNodeType struct {
	Type     NodeType
	Name     string
	Children []Child
}

type CloneableInterface struct {
	Name       string
	UniqueFunc string
	Structs    []string
}

type Child struct {
	FieldName string
	FieldType string
	Cloneable bool
	Pointer   bool
	Optional  bool
	Interface *CloneableInterface
}

func newChild(fieldName, fieldType string, cloneable, pointer, optional bool) Child {
	return Child{FieldName: fieldName, FieldType: fieldType, Cloneable: cloneable, Pointer: pointer, Optional: optional}
}

func main() {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "./ast", func(info fs.FileInfo) bool {
		return info.Name() != "clone.go" && info.Name() != "visit.go"
	}, parser.ParseComments)
	if err != nil {
		log.Fatalf("%v", err)
	}

	var nodes []CloneableNodeType
	var interfaces []CloneableInterface
	for _, file := range pkgs["ast"].Files {
		nodes = append(nodes, findCloneableNodes(file)...)
		interfaces = append(interfaces, findCloneableInterfaces(file)...)
	}
	for _, file := range pkgs["ast"].Files {
		findStructsForInterfaces(file, interfaces)
	}

	slices.SortFunc(nodes, func(a, b CloneableNodeType) int {
		return cmp.Compare(a.Name, b.Name)
	})
	for i := range nodes {
		for j := range nodes[i].Children {
			if idx := slices.IndexFunc(interfaces, func(a CloneableInterface) bool {
				return a.Name == nodes[i].Children[j].FieldType
			}); idx != -1 {
				nodes[i].Children[j].Interface = &interfaces[idx]
			}
		}
	}
	for _, i := range interfaces {
		slices.SortFunc(i.Structs, func(a, b string) int {
			return cmp.Compare(a, b)
		})
	}
	fmt.Println(nodes)

	var (
		visitMethods []ast.Decl
	)
	for _, node := range nodes {
		recv := newFieldList("n", &ast.StarExpr{X: ast.NewIdent(node.Name)})
		visitChildrenBlock := &ast.BlockStmt{}

		var fields []ast.Expr
		switch node.Type {
		case NodeTypeStruct:
			for _, child := range node.Children {
				if !child.Cloneable {
					fields = append(fields, &ast.KeyValueExpr{
						Key:   ast.NewIdent(child.FieldName),
						Value: newSelectorExpr(ast.NewIdent("n"), child.FieldName),
					})
					continue
				}
				if child.Interface != nil {
					declStmt, switchStmt := newCloner(newSelectorExpr(ast.NewIdent("n"), child.FieldName), child.Interface)
					visitChildrenBlock.List = append(visitChildrenBlock.List, declStmt, switchStmt)

					fields = append(fields, &ast.KeyValueExpr{
						Key:   ast.NewIdent(child.FieldName),
						Value: declStmt.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec).Names[0],
					})
					continue
				}
				if child.Optional {
					visitChildrenBlock.List = append(visitChildrenBlock.List,
						&ast.DeclStmt{
							Decl: &ast.GenDecl{
								Tok: token.VAR,
								Specs: []ast.Spec{
									&ast.ValueSpec{
										Names: []*ast.Ident{ast.NewIdent(strings.ToLower(child.FieldName))},
										Type:  &ast.StarExpr{X: ast.NewIdent(child.FieldType)},
									},
								},
							},
						}, &ast.IfStmt{
							Cond: &ast.BinaryExpr{
								X:  newSelectorExpr(ast.NewIdent("n"), child.FieldName),
								Op: token.NEQ,
								Y:  ast.NewIdent("nil"),
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.AssignStmt{
										Lhs: []ast.Expr{ast.NewIdent(strings.ToLower(child.FieldName))},
										Tok: token.ASSIGN,
										Rhs: []ast.Expr{&ast.CallExpr{
											Fun: newSelectorExpr(
												newSelectorExpr(ast.NewIdent("n"), child.FieldName),
												"Clone",
											),
										}},
									},
								},
							},
						})
					fields = append(fields, &ast.KeyValueExpr{
						Key:   ast.NewIdent(child.FieldName),
						Value: ast.NewIdent(strings.ToLower(child.FieldName)),
					})
					continue
				}
				if child.Pointer {
					fields = append(fields, &ast.KeyValueExpr{
						Key: ast.NewIdent(child.FieldName),
						Value: &ast.CallExpr{
							Fun: newSelectorExpr(
								newSelectorExpr(ast.NewIdent("n"), child.FieldName),
								"Clone",
							),
						},
					})
					continue
				}
				fields = append(fields, &ast.KeyValueExpr{
					Key: ast.NewIdent(child.FieldName),
					Value: &ast.StarExpr{X: &ast.CallExpr{
						Fun: newSelectorExpr(
							newSelectorExpr(ast.NewIdent("n"), child.FieldName),
							"Clone",
						),
					}},
				})
			}

			visitChildrenBlock.List = append(visitChildrenBlock.List, &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.UnaryExpr{Op: token.AND, X: &ast.CompositeLit{
						Type: ast.NewIdent(node.Name),
						Elts: fields,
					}},
				},
			})
		case NodeTypeSlice:
			// for i := range *n {
			//     (*n)[i].VisitWith(v)
			// }
			visitChildrenBlock.List = append(visitChildrenBlock.List, &ast.AssignStmt{
				Lhs: []ast.Expr{ast.NewIdent("ns")},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{&ast.CallExpr{
					Fun: ast.NewIdent("make"),
					Args: []ast.Expr{ast.NewIdent(node.Name), &ast.CallExpr{
						Fun:  ast.NewIdent("len"),
						Args: []ast.Expr{&ast.StarExpr{X: ast.NewIdent("n")}},
					}},
				}},
			}, &ast.RangeStmt{
				Key: ast.NewIdent("i"),
				Tok: token.DEFINE,
				X:   &ast.StarExpr{X: ast.NewIdent("n")},
				Body: &ast.BlockStmt{
					List: []ast.Stmt{
						&ast.AssignStmt{Lhs: []ast.Expr{&ast.IndexExpr{
							X:     &ast.Ident{Name: "ns"},
							Index: ast.NewIdent("i"),
						}}, Tok: token.ASSIGN, Rhs: []ast.Expr{&ast.StarExpr{X: &ast.CallExpr{
							Fun: newSelectorExpr(&ast.IndexExpr{
								X:     &ast.StarExpr{X: ast.NewIdent("n")},
								Index: ast.NewIdent("i"),
							}, "Clone"),
						}}}},
					},
				},
			}, &ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.UnaryExpr{Op: token.AND, X: ast.NewIdent("ns")},
				},
			})
		}

		visitMethods = append(visitMethods, &ast.FuncDecl{
			Recv: recv,
			Name: ast.NewIdent("Clone"),
			Type: &ast.FuncType{Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{X: &ast.Ident{Name: node.Name}},
					},
				},
			}},
			Body: visitChildrenBlock,
		})
	}

	genPkg := &ast.File{
		Name: ast.NewIdent("ast"),
	}

	genPkg.Decls = append(genPkg.Decls, visitMethods...)

	s := bytes.NewBuffer([]byte("// Code generated by gen_clone.go; DO NOT EDIT.\n"))
	format.Node(s, fset, genPkg)

	os.WriteFile("ast/clone.go", s.Bytes(), 0644)

	fmt.Println(pkgs)
}

func findCloneableInterfaces(f *ast.File) []CloneableInterface {
	var interfaces []CloneableInterface
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			switch typeSpec.Name.Name {
			case "Node", "VisitableNode", "CloneableNode":
				continue
			}

			switch t := typeSpec.Type.(type) {
			case *ast.InterfaceType:
				idx := slices.IndexFunc(t.Methods.List, func(a *ast.Field) bool {
					if len(a.Names) == 0 {
						return false
					}
					return strings.HasPrefix(a.Names[0].Name, "_")
				})
				if idx == -1 {
					continue
				}
				interfaces = append(interfaces, CloneableInterface{
					Name:       typeSpec.Name.Name,
					UniqueFunc: t.Methods.List[idx].Names[0].Name,
				})
			}
		}
	}
	return interfaces
}

func findStructsForInterfaces(f *ast.File, interfaces []CloneableInterface) {
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}
		starExpr, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		ident, ok := starExpr.X.(*ast.Ident)
		if !ok {
			continue
		}
		idx := slices.IndexFunc(interfaces, func(a CloneableInterface) bool {
			return a.UniqueFunc == funcDecl.Name.Name
		})
		if idx == -1 {
			continue
		}
		interfaces[idx].Structs = append(interfaces[idx].Structs, ident.Name)
	}
}

func findCloneableNodes(f *ast.File) (types []CloneableNodeType) {
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			switch typeSpec.Name.Name {
			case "ScopeContext", "Id":
				continue
			}

			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				types = append(types, CloneableNodeType{
					Type:     NodeTypeStruct,
					Name:     typeSpec.Name.Name,
					Children: findStructChildren(t.Fields.List),
				})
			case *ast.ArrayType:
				types = append(types, CloneableNodeType{
					Type: NodeTypeSlice,
					Name: typeSpec.Name.Name,
				})
			}
		}
	}
	return types
}

func findStructChildren(fields []*ast.Field) (children []Child) {
	for _, field := range fields {
		optional := field.Tag != nil && field.Tag.Value == "`optional:\"true\"`"
		if len(field.Names) != 0 {
			fmt.Println(field.Names[0].Name)
		}

		switch fieldType := field.Type.(type) {
		case *ast.SelectorExpr:
			children = append(children, newChild(field.Names[0].Name, "", false, false, optional))
		case *ast.Ident:
			if len(field.Names) == 0 {
				children = append(children, newChild(fieldType.Name, fieldType.Name, true, false, optional))
				continue
			}

			switch fieldType.Name {
			case "Idx", "any", "bool", "int", "ScopeContext", "string", "PropertyKind", "Token", "float64":
				children = append(children, newChild(field.Names[0].Name, fieldType.Name, false, false, optional))
			default:
				children = append(children, newChild(field.Names[0].Name, fieldType.Name, true, false, optional))
			}
		case *ast.StarExpr:
			if ident, ok := fieldType.X.(*ast.Ident); ok {
				if ident.Name == "string" {
					children = append(children, newChild(field.Names[0].Name, ident.Name, false, false, optional))
					continue
				}
				children = append(children, newChild(field.Names[0].Name, ident.Name, true, true, optional))
			} else {
				children = append(children, newChild(field.Names[0].Name, "", true, false, optional))
			}
		}
	}
	return children
}

func newFieldList(name string, t ast.Expr) *ast.FieldList {
	return &ast.FieldList{
		List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  t,
		}},
	}
}

func newSelectorExpr(x ast.Expr, sel string) *ast.SelectorExpr {
	return &ast.SelectorExpr{X: x, Sel: ast.NewIdent(sel)}
}

func lowerIdent(name string) *ast.Ident {
	return ast.NewIdent(strings.ToLower(name[:1]) + name[1:])
}

func newCloner(expr ast.Expr, intf *CloneableInterface) (*ast.DeclStmt, *ast.TypeSwitchStmt) {
	clonedExprDecl := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{ast.NewIdent("cloned" + intf.Name)},
					Type:  ast.NewIdent(intf.Name),
				},
			},
		},
	}

	var cases []ast.Stmt

	for _, structName := range intf.Structs {
		caseClause := &ast.CaseClause{
			List: []ast.Expr{
				&ast.StarExpr{X: ast.NewIdent(structName)},
			},
			Body: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("cloned" + intf.Name)},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   lowerIdent(intf.Name),
								Sel: ast.NewIdent("Clone"),
							},
						},
					},
				},
			},
		}
		cases = append(cases, caseClause)
	}

	switchStmt := &ast.TypeSwitchStmt{
		Assign: &ast.AssignStmt{
			Lhs: []ast.Expr{
				lowerIdent(intf.Name),
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.TypeAssertExpr{
					X: expr,
				},
			},
		},
		Body: &ast.BlockStmt{
			List: cases,
		},
	}

	return clonedExprDecl, switchStmt
}
