package process

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

func File(af *ast.File, fSet *token.FileSet, commentBase, commentIndentedBase, template string, parenComment bool) (_ *ast.File, modified bool, err error) {
	// Inject first comment to prevent nil comment map
	if len(af.Comments) == 0 {
		af.Comments = []*ast.CommentGroup{{List: []*ast.Comment{{Slash: -1, Text: "// gocmt"}}}}
		defer func() {
			// Remove the injected comment
			af.Comments = af.Comments[1:]
		}()
	}

	commentTemplate := commentBase + template
	numComments := len(af.Comments)
	cMap := ast.NewCommentMap(fSet, af, af.Comments)

	skipped := make(map[ast.Node]bool)
	ast.Inspect(af, func(n ast.Node) bool {
		switch typ := n.(type) {
		case *ast.FuncDecl:
			if skipped[typ] || !typ.Name.IsExported() {
				return true
			}
			addFuncDeclComment(typ, commentTemplate)
			cMap[typ] = []*ast.CommentGroup{typ.Doc}

		case *ast.DeclStmt:
			skipped[typ.Decl] = true

		case *ast.GenDecl:
			switch typ.Tok {
			case token.CONST, token.VAR:
				if !(typ.Lparen == token.NoPos && typ.Rparen == token.NoPos) {
					// if there's a () and parenComment is true, add comment for each sub entry
					if parenComment {
						for _, spec := range typ.Specs {
							vs := spec.(*ast.ValueSpec)
							if !vs.Names[0].IsExported() {
								continue
							}
							addParenValueSpecComment(vs, commentTemplate, commentBase, commentIndentedBase)
							cMap[vs] = []*ast.CommentGroup{vs.Doc}
						}
						return true
					}
				}

				// empty var block
				if len(typ.Specs) == 0 {
					return true
				}

				vs := typ.Specs[0].(*ast.ValueSpec)
				if skipped[typ] || !vs.Names[0].IsExported() {
					return true
				}
				addValueSpecComment(typ, vs, commentTemplate)

			case token.TYPE:
				ts := typ.Specs[0].(*ast.TypeSpec)
				if skipped[typ] || !ts.Name.IsExported() {
					return true
				}
				addTypeSpecComment(typ, ts, commentTemplate)
			default:
				return true
			}
			cMap[typ] = []*ast.CommentGroup{typ.Doc}
		}
		return true
	})

	// Rebuild comments
	af.Comments = cMap.Filter(af).Comments()
	modified = len(af.Comments) > numComments
	return
}
func addFuncDeclComment(fd *ast.FuncDecl, commentTemplate string) {
	if fd.Doc == nil || strings.TrimSpace(fd.Doc.Text()) == fd.Name.Name {
		text := fmt.Sprintf(commentTemplate, fd.Name)
		pos := fd.Pos() - token.Pos(1)
		if fd.Doc != nil {
			pos = fd.Doc.Pos()
		}
		fd.Doc = &ast.CommentGroup{List: []*ast.Comment{{Slash: pos, Text: text}}}
	}

}

func addValueSpecComment(gd *ast.GenDecl, vs *ast.ValueSpec, commentTemplate string) {
	if gd.Doc == nil || strings.TrimSpace(gd.Doc.Text()) == vs.Names[0].Name {
		text := fmt.Sprintf(commentTemplate, vs.Names[0].Name)
		pos := gd.Pos() - token.Pos(1)
		if gd.Doc != nil {
			pos = gd.Doc.Pos()
		}
		gd.Doc = &ast.CommentGroup{List: []*ast.Comment{{Slash: pos, Text: text}}}
	}
}

func addParenValueSpecComment(vs *ast.ValueSpec, commentTemplate, commentBase, commentIndentedBase string) {
	if vs.Doc == nil || strings.TrimSpace(vs.Doc.Text()) == vs.Names[0].Name {
		commentTemplate = strings.Replace(commentTemplate, commentBase, commentIndentedBase, 1)
		text := fmt.Sprintf(commentTemplate, vs.Names[0].Name)
		pos := vs.Pos() - token.Pos(1)
		if vs.Doc != nil {
			pos = vs.Doc.Pos()
		}
		vs.Doc = &ast.CommentGroup{List: []*ast.Comment{{Slash: pos, Text: text}}}
	}
}

func addTypeSpecComment(gd *ast.GenDecl, ts *ast.TypeSpec, commentTemplate string) {
	if gd.Doc == nil || strings.TrimSpace(gd.Doc.Text()) == ts.Name.Name {
		text := fmt.Sprintf(commentTemplate, ts.Name.Name)
		pos := gd.Pos() - token.Pos(1)
		if gd.Doc != nil {
			pos = gd.Doc.Pos()
		}
		gd.Doc = &ast.CommentGroup{List: []*ast.Comment{{Slash: pos, Text: text}}}
	}
}
