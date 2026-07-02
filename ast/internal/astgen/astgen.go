// Package astgen discovers the tagged-union specs shared by the ast code
// generators (gen_union.go, gen_visit.go, gen_clone.go).
package astgen

import (
	"cmp"
	"go/ast"
	"go/token"
	"slices"
	"strings"
)

// UnionVariant is one member of a tagged union.
type UnionVariant struct {
	TypeName  string // payload struct, e.g. "BinaryExpression"
	ShortName string // tag used in Kind constants/accessors, e.g. "Binary"
}

// UnionSpec describes one tagged union, discovered from a //union: spec struct.
type UnionSpec struct {
	WrapperType    string // e.g. "Expression"
	KindType       string // e.g. "ExprKind"
	KindPrefix     string // e.g. "Expr"
	ConstructorSfx string // e.g. "Expr"
	Variants       []UnionVariant
}

// UnionWrapper returns the type named in a `//union:Wrapper` doc comment, or ""
// if absent. Such structs are union specs, not real AST nodes.
func UnionWrapper(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	for _, c := range doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(c.Text, "//"))
		if rest, ok := strings.CutPrefix(text, "union:"); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// FindUnionSpecs returns the //union: specs keyed by wrapper type. Each spec
// field is a variant (name -> ShortName, type -> TypeName), sorted for stable
// output.
func FindUnionSpecs(files map[string]*ast.File) map[string]UnionSpec {
	specs := map[string]UnionSpec{}
	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				wrapper := UnionWrapper(typeSpec.Doc)
				if wrapper == "" {
					continue
				}

				kindPrefix, kindType, ctorSuffix := DeriveUnionNames(wrapper)
				var variants []UnionVariant
				for _, field := range structType.Fields.List {
					ident, ok := field.Type.(*ast.Ident)
					if !ok {
						continue
					}
					for _, name := range field.Names {
						variants = append(variants, UnionVariant{
							TypeName:  ident.Name,
							ShortName: name.Name,
						})
					}
				}
				slices.SortFunc(variants, func(a, b UnionVariant) int {
					return cmp.Compare(a.ShortName, b.ShortName)
				})

				specs[wrapper] = UnionSpec{
					WrapperType:    wrapper,
					KindType:       kindType,
					KindPrefix:     kindPrefix,
					ConstructorSfx: ctorSuffix,
					Variants:       variants,
				}
			}
		}
	}
	return specs
}

// DeriveUnionNames maps a wrapper type to its Kind prefix, Kind type, and
// constructor suffix.
func DeriveUnionNames(name string) (kindPrefix, kindType, ctorSuffix string) {
	switch name {
	case "Expression":
		return "Expr", "ExprKind", "Expr"
	case "Statement":
		return "Stmt", "StmtKind", "Stmt"
	case "Property":
		return "Prop", "PropKind", "Prop"
	case "MemberProperty":
		return "MemProp", "MemPropKind", "MemProp"
	case "ClassElement":
		return "ClassElem", "ClassElemKind", "ClassElem"
	case "PatternProperty":
		return "PatProp", "PatPropKind", "PatProp"
	case "PropertyName":
		return "PropName", "PropNameKind", "PropName"
	default:
		return name, name + "Kind", name
	}
}
