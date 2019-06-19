// Warnings about deprecated operations in Starlark

package warn

import (
	"github.com/bazelbuild/buildtools/build"
)

func dictionaryConcatenationWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	var addWarning = func(expr build.Expr) {
		findings = append(findings,
			makeLinterFinding(expr, "Dictionary concatenation is deprecated."))
	}

	types := detectTypes(f)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.BinaryExpr:
			if expr.Op != "+" {
				return
			}
			if types[expr.X] == Dict || types[expr.Y] == Dict {
				addWarning(expr)
			}
		case *build.AssignExpr:
			if expr.Op != "+=" {
				return
			}
			if types[expr.LHS] == Dict || types[expr.RHS] == Dict {
				addWarning(expr)
			}
		}
	})
	return findings
}

func stringIterationWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	addWarning := func(expr build.Expr) {
		findings = append(findings,
			makeLinterFinding(expr, "String iteration is deprecated."))
	}

	types := detectTypes(f)
	build.Walk(f, func(expr build.Expr, stack []build.Expr) {
		switch expr := expr.(type) {
		case *build.ForStmt:
			if types[expr.X] == String {
				addWarning(expr.X)
			}
		case *build.ForClause:
			if types[expr.X] == String {
				addWarning(expr.X)
			}
		case *build.CallExpr:
			ident, ok := expr.X.(*build.Ident)
			if !ok {
				return
			}
			switch ident.Name {
			case "all", "any", "reversed", "max", "min":
				if len(expr.List) != 1 {
					return
				}
				if types[expr.List[0]] == String {
					addWarning(expr.List[0])
				}
			case "zip":
				for _, arg := range expr.List {
					if types[arg] == String {
						addWarning(arg)
					}
				}
			}
		}
	})
	return findings
}

func integerDivisionWarning(f *build.File) []*LinterFinding {
	var findings []*LinterFinding

	build.WalkPointers(f, func(e *build.Expr, stack []build.Expr) {
		switch expr := (*e).(type) {
		case *build.BinaryExpr:
			if expr.Op != "/" {
				return
			}
			newBinary := *expr
			newBinary.Op = "//"
			findings = append(findings,
				makeLinterFinding(expr, `The "/" operator for integer division is deprecated in favor of "//".`,
					LinterReplacement{e, &newBinary}))

		case *build.AssignExpr:
			if expr.Op != "/=" {
				return
			}
			newAssign := *expr
			newAssign.Op = "//="
			findings = append(findings,
				makeLinterFinding(expr, `The "/=" operator for integer division is deprecated in favor of "//=".`,
					LinterReplacement{e, &newAssign}))
		}
	})
	return findings
}
