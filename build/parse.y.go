//line build/parse.y:13
package build

import __yyfmt__ "fmt"

//line build/parse.y:13

//line build/parse.y:18
type yySymType struct {
	yys int
	// input tokens
	tok    string   // raw input syntax
	str    string   // decoding of quoted string
	pos    Position // position of token
	triple bool     // was string triple quoted?

	// partial syntax trees
	expr    Expr
	exprs   []Expr
	kv      *KeyValueExpr
	kvs     []*KeyValueExpr
	string  *StringExpr
	ifstmt  *IfStmt
	loadarg *struct {
		from Ident
		to   Ident
	}
	loadargs []*struct {
		from Ident
		to   Ident
	}

	// supporting information
	comma    Position // position of trailing comma in list, if present
	lastStmt Expr     // most recent rule, to attach line comments to
}

const _AUGM = 57346
const _AND = 57347
const _COMMENT = 57348
const _EOF = 57349
const _EQ = 57350
const _FOR = 57351
const _GE = 57352
const _IDENT = 57353
const _INT = 57354
const _IF = 57355
const _ELSE = 57356
const _ELIF = 57357
const _IN = 57358
const _IS = 57359
const _LAMBDA = 57360
const _LOAD = 57361
const _LE = 57362
const _NE = 57363
const _STAR_STAR = 57364
const _INT_DIV = 57365
const _BIT_LSH = 57366
const _BIT_RSH = 57367
const _NOT = 57368
const _OR = 57369
const _STRING = 57370
const _DEF = 57371
const _RETURN = 57372
const _PASS = 57373
const _BREAK = 57374
const _CONTINUE = 57375
const _INDENT = 57376
const _UNINDENT = 57377
const ShiftInstead = 57378
const _ASSERT = 57379
const _UNARY = 57380

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"'%'",
	"'('",
	"')'",
	"'*'",
	"'+'",
	"','",
	"'-'",
	"'.'",
	"'/'",
	"':'",
	"'<'",
	"'='",
	"'>'",
	"'['",
	"']'",
	"'{'",
	"'}'",
	"'|'",
	"'&'",
	"'^'",
	"'~'",
	"_AUGM",
	"_AND",
	"_COMMENT",
	"_EOF",
	"_EQ",
	"_FOR",
	"_GE",
	"_IDENT",
	"_INT",
	"_IF",
	"_ELSE",
	"_ELIF",
	"_IN",
	"_IS",
	"_LAMBDA",
	"_LOAD",
	"_LE",
	"_NE",
	"_STAR_STAR",
	"_INT_DIV",
	"_BIT_LSH",
	"_BIT_RSH",
	"_NOT",
	"_OR",
	"_STRING",
	"_DEF",
	"_RETURN",
	"_PASS",
	"_BREAK",
	"_CONTINUE",
	"_INDENT",
	"_UNINDENT",
	"ShiftInstead",
	"'\\n'",
	"_ASSERT",
	"_UNARY",
	"';'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line build/parse.y:975

// Go helper code.

// unary returns a unary expression with the given
// position, operator, and subexpression.
func unary(pos Position, op string, x Expr) Expr {
	return &UnaryExpr{
		OpStart: pos,
		Op:      op,
		X:       x,
	}
}

// binary returns a binary expression with the given
// operands, position, and operator.
func binary(x Expr, pos Position, op string, y Expr) Expr {
	_, xend := x.Span()
	ystart, _ := y.Span()

	switch op {
	case "=", "+=", "-=", "*=", "/=", "//=", "%=", "|=":
		return &AssignExpr{
			LHS:       x,
			OpPos:     pos,
			Op:        op,
			LineBreak: xend.Line < ystart.Line,
			RHS:       y,
		}
	}

	return &BinaryExpr{
		X:         x,
		OpStart:   pos,
		Op:        op,
		LineBreak: xend.Line < ystart.Line,
		Y:         y,
	}
}

// isSimpleExpression returns whether an expression is simple and allowed to exist in
// compact forms of sequences.
// The formal criteria are the following: an expression is considered simple if it's
// a literal (variable, string or a number), a literal with a unary operator or an empty sequence.
func isSimpleExpression(expr *Expr) bool {
	switch x := (*expr).(type) {
	case *LiteralExpr, *StringExpr, *Ident:
		return true
	case *UnaryExpr:
		_, literal := x.X.(*LiteralExpr)
		_, ident := x.X.(*Ident)
		return literal || ident
	case *ListExpr:
		return len(x.List) == 0
	case *TupleExpr:
		return len(x.List) == 0
	case *DictExpr:
		return len(x.List) == 0
	case *SetExpr:
		return len(x.List) == 0
	default:
		return false
	}
}

// forceCompact returns the setting for the ForceCompact field for a call or tuple.
//
// NOTE 1: The field is called ForceCompact, not ForceSingleLine,
// because it only affects the formatting associated with the call or tuple syntax,
// not the formatting of the arguments. For example:
//
//	call([
//		1,
//		2,
//		3,
//	])
//
// is still a compact call even though it runs on multiple lines.
//
// In contrast the multiline form puts a linebreak after the (.
//
//	call(
//		[
//			1,
//			2,
//			3,
//		],
//	)
//
// NOTE 2: Because of NOTE 1, we cannot use start and end on the
// same line as a signal for compact mode: the formatting of an
// embedded list might move the end to a different line, which would
// then look different on rereading and cause buildifier not to be
// idempotent. Instead, we have to look at properties guaranteed
// to be preserved by the reformatting, namely that the opening
// paren and the first expression are on the same line and that
// each subsequent expression begins on the same line as the last
// one ended (no line breaks after comma).
func forceCompact(start Position, list []Expr, end Position) bool {
	if len(list) <= 1 {
		// The call or tuple will probably be compact anyway; don't force it.
		return false
	}

	// If there are any named arguments or non-string, non-literal
	// arguments, cannot force compact mode.
	line := start.Line
	for _, x := range list {
		start, end := x.Span()
		if start.Line != line {
			return false
		}
		line = end.Line
		if !isSimpleExpression(&x) {
			return false
		}
	}
	return end.Line == line
}

// forceMultiLine returns the setting for the ForceMultiLine field.
func forceMultiLine(start Position, list []Expr, end Position) bool {
	if len(list) > 1 {
		// The call will be multiline anyway, because it has multiple elements. Don't force it.
		return false
	}

	if len(list) == 0 {
		// Empty list: use position of brackets.
		return start.Line != end.Line
	}

	// Single-element list.
	// Check whether opening bracket is on different line than beginning of
	// element, or closing bracket is on different line than end of element.
	elemStart, elemEnd := list[0].Span()
	return start.Line != elemStart.Line || end.Line != elemEnd.Line
}

// forceMultiLineComprehension returns the setting for the ForceMultiLine field for a comprehension.
func forceMultiLineComprehension(start Position, expr Expr, clauses []Expr, end Position) bool {
	// Return true if there's at least one line break between start, expr, each clause, and end
	exprStart, exprEnd := expr.Span()
	if start.Line != exprStart.Line {
		return true
	}
	previousEnd := exprEnd
	for _, clause := range clauses {
		clauseStart, clauseEnd := clause.Span()
		if previousEnd.Line != clauseStart.Line {
			return true
		}
		previousEnd = clauseEnd
	}
	return previousEnd.Line != end.Line
}

// extractTrailingComments extracts trailing comments of an indented block starting with the first
// comment line with indentation less than the block indentation.
// The comments can either belong to CommentBlock statements or to the last non-comment statement
// as After-comments.
func extractTrailingComments(stmt Expr) []Expr {
	body := getLastBody(stmt)
	var comments []Expr
	if body != nil && len(*body) > 0 {
		// Get the current indentation level
		start, _ := (*body)[0].Span()
		indentation := start.LineRune

		// Find the last non-comment statement
		lastNonCommentIndex := -1
		for i, stmt := range *body {
			if _, ok := stmt.(*CommentBlock); !ok {
				lastNonCommentIndex = i
			}
		}
		if lastNonCommentIndex == -1 {
			return comments
		}

		// Iterate over the trailing comments, find the first comment line that's not indented enough,
		// dedent it and all the following comments.
		for i := lastNonCommentIndex; i < len(*body); i++ {
			stmt := (*body)[i]
			if comment := extractDedentedComment(stmt, indentation); comment != nil {
				// This comment and all the following CommentBlock statements are to be extracted.
				comments = append(comments, comment)
				comments = append(comments, (*body)[i+1:]...)
				*body = (*body)[:i+1]
				// If the current statement is a CommentBlock statement without any comment lines
				// it should be removed too.
				if i > lastNonCommentIndex && len(stmt.Comment().After) == 0 {
					*body = (*body)[:i]
				}
			}
		}
	}
	return comments
}

// extractDedentedComment extract the first comment line from `stmt` which indentation is smaller
// than `indentation`, and all following comment lines, and returns them in a newly created
// CommentBlock statement.
func extractDedentedComment(stmt Expr, indentation int) Expr {
	for i, line := range stmt.Comment().After {
		// line.Start.LineRune == 0 can't exist in parsed files, it indicates that the comment line
		// has been added by an AST modification. Don't take such lines into account.
		if line.Start.LineRune > 0 && line.Start.LineRune < indentation {
			// This and all the following lines should be dedented
			cb := &CommentBlock{
				Start:    line.Start,
				Comments: Comments{After: stmt.Comment().After[i:]},
			}
			stmt.Comment().After = stmt.Comment().After[:i]
			return cb
		}
	}
	return nil
}

// getLastBody returns the last body of a block statement (the only body for For- and DefStmt
// objects, the last in a if-elif-else chain
func getLastBody(stmt Expr) *[]Expr {
	switch block := stmt.(type) {
	case *DefStmt:
		return &block.Body
	case *ForStmt:
		return &block.Body
	case *IfStmt:
		if len(block.False) == 0 {
			return &block.True
		} else if len(block.False) == 1 {
			if next, ok := block.False[0].(*IfStmt); ok {
				// Recursively find the last block of the chain
				return getLastBody(next)
			}
		}
		return &block.False
	}
	return nil
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 861

var yyAct = [...]int{

	19, 173, 27, 214, 7, 164, 29, 2, 144, 135,
	211, 148, 171, 88, 9, 41, 21, 98, 147, 224,
	235, 226, 160, 38, 80, 81, 82, 83, 42, 37,
	121, 34, 86, 91, 94, 51, 101, 51, 50, 54,
	50, 55, 190, 52, 159, 52, 104, 223, 37, 166,
	225, 108, 109, 110, 111, 112, 113, 114, 115, 116,
	117, 118, 119, 120, 218, 122, 123, 124, 125, 126,
	127, 128, 129, 130, 34, 53, 136, 53, 137, 48,
	49, 33, 101, 13, 25, 167, 24, 36, 96, 34,
	106, 145, 132, 31, 40, 32, 152, 87, 46, 155,
	26, 85, 165, 194, 73, 188, 79, 153, 34, 35,
	100, 107, 150, 161, 193, 22, 30, 150, 169, 154,
	33, 241, 174, 23, 150, 37, 36, 102, 103, 146,
	90, 105, 31, 228, 32, 93, 231, 45, 227, 181,
	182, 201, 180, 178, 176, 45, 183, 34, 35, 45,
	75, 43, 192, 204, 219, 30, 74, 195, 197, 189,
	140, 44, 76, 184, 37, 189, 191, 42, 142, 202,
	203, 196, 199, 45, 45, 230, 208, 170, 131, 95,
	136, 210, 137, 205, 168, 212, 200, 187, 179, 209,
	216, 217, 45, 157, 215, 151, 143, 237, 198, 177,
	158, 99, 221, 165, 84, 97, 220, 51, 207, 1,
	50, 54, 213, 55, 206, 52, 92, 222, 89, 232,
	39, 47, 229, 18, 12, 67, 185, 186, 8, 212,
	4, 234, 233, 216, 238, 28, 7, 215, 239, 236,
	33, 149, 77, 25, 78, 24, 36, 53, 69, 70,
	141, 162, 31, 163, 32, 133, 134, 0, 0, 26,
	0, 0, 6, 0, 0, 11, 0, 34, 35, 20,
	0, 0, 0, 0, 22, 30, 0, 0, 0, 0,
	0, 0, 23, 0, 37, 10, 14, 15, 16, 17,
	0, 240, 33, 5, 0, 25, 0, 24, 36, 0,
	0, 0, 0, 0, 31, 0, 32, 0, 0, 0,
	0, 26, 0, 0, 6, 3, 0, 11, 0, 34,
	35, 20, 0, 0, 0, 51, 22, 30, 50, 54,
	0, 55, 0, 52, 23, 0, 37, 10, 14, 15,
	16, 17, 0, 33, 0, 5, 25, 0, 24, 36,
	0, 0, 0, 0, 0, 31, 0, 32, 0, 0,
	0, 0, 26, 0, 0, 53, 69, 70, 0, 0,
	34, 35, 0, 0, 0, 0, 0, 22, 30, 0,
	0, 0, 0, 0, 0, 23, 0, 37, 0, 14,
	15, 16, 17, 0, 51, 0, 172, 50, 54, 0,
	55, 0, 52, 156, 56, 0, 57, 0, 0, 0,
	0, 66, 67, 68, 0, 0, 65, 0, 0, 58,
	0, 61, 0, 0, 72, 0, 0, 62, 71, 0,
	0, 59, 60, 0, 53, 69, 70, 63, 64, 51,
	0, 0, 50, 54, 0, 55, 0, 52, 0, 56,
	0, 57, 0, 0, 0, 0, 66, 67, 68, 0,
	0, 65, 0, 0, 58, 0, 61, 0, 0, 72,
	175, 0, 62, 71, 0, 0, 59, 60, 0, 53,
	69, 70, 63, 64, 51, 0, 0, 50, 54, 0,
	55, 0, 52, 0, 56, 0, 57, 0, 0, 0,
	0, 66, 67, 68, 0, 0, 65, 0, 0, 58,
	150, 61, 0, 0, 72, 0, 0, 62, 71, 0,
	0, 59, 60, 0, 53, 69, 70, 63, 64, 51,
	0, 0, 50, 54, 0, 55, 0, 52, 0, 56,
	0, 57, 0, 0, 0, 0, 66, 67, 68, 0,
	0, 65, 0, 0, 58, 0, 61, 0, 0, 72,
	0, 0, 62, 71, 0, 0, 59, 60, 0, 53,
	69, 70, 63, 64, 33, 0, 0, 25, 0, 24,
	36, 0, 0, 0, 0, 0, 31, 0, 32, 0,
	0, 0, 0, 26, 0, 0, 0, 0, 0, 0,
	0, 34, 35, 0, 0, 0, 0, 51, 22, 30,
	50, 54, 0, 55, 0, 52, 23, 56, 37, 57,
	14, 15, 16, 17, 66, 67, 68, 0, 0, 65,
	0, 0, 58, 0, 61, 0, 0, 0, 0, 0,
	62, 71, 0, 0, 59, 60, 0, 53, 69, 70,
	63, 64, 51, 0, 0, 50, 54, 0, 55, 0,
	52, 0, 56, 0, 57, 0, 0, 0, 0, 66,
	67, 68, 0, 0, 65, 0, 0, 58, 0, 61,
	0, 0, 0, 0, 0, 62, 0, 0, 0, 59,
	60, 0, 53, 69, 70, 63, 64, 51, 0, 0,
	50, 54, 0, 55, 0, 52, 0, 56, 0, 57,
	0, 0, 0, 0, 66, 67, 68, 0, 0, 65,
	0, 0, 58, 0, 61, 0, 0, 0, 0, 0,
	62, 0, 0, 0, 59, 60, 0, 53, 69, 70,
	63, 51, 0, 0, 50, 54, 0, 55, 0, 52,
	0, 56, 0, 57, 0, 0, 0, 0, 66, 67,
	68, 0, 0, 0, 0, 0, 58, 0, 61, 0,
	0, 0, 0, 0, 62, 0, 0, 0, 59, 60,
	0, 53, 69, 70, 63, 33, 0, 138, 25, 0,
	24, 36, 51, 0, 0, 50, 54, 31, 55, 32,
	52, 0, 0, 0, 26, 0, 0, 0, 0, 66,
	67, 68, 34, 35, 0, 0, 0, 0, 51, 22,
	30, 50, 54, 139, 55, 0, 52, 23, 0, 37,
	0, 0, 53, 69, 70, 0, 67, 68, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 53, 69,
	70,
}
var yyPact = [...]int{

	-1000, -1000, 287, -1000, -1000, -1000, -35, -1000, -1000, -1000,
	62, 115, -1000, 136, 76, -1000, -1000, -1000, 44, 525,
	76, 145, 76, 76, 76, 76, 76, -1000, -1000, -1000,
	199, 76, 76, 76, -1000, 168, 55, -1000, -1000, -44,
	196, 73, 145, 76, 76, 76, 183, 76, 77, -1000,
	76, 76, 76, 76, 76, 76, 76, 76, 76, 76,
	76, 76, 76, -7, 76, 76, 76, 76, 76, 76,
	76, 76, 76, 165, 60, 780, 76, 155, 187, 183,
	-1000, -1000, -1000, -1000, -20, 111, 480, 186, 82, 99,
	186, 390, 184, 194, 525, 11, -1000, -36, 569, 42,
	76, 115, 183, 183, 603, 164, 338, -1000, -1000, -1000,
	-1000, -1000, 33, 33, 788, 788, 788, 788, 788, 788,
	788, 76, 693, 737, 814, 321, 203, 31, 31, 648,
	435, 338, -1000, 193, 179, -1000, 525, 127, 76, 76,
	128, 150, 76, 76, -1000, 178, -1000, 87, 8, -1000,
	115, 76, -1000, 94, -1000, 83, 76, 76, -1000, -1000,
	-1000, -1000, 192, 177, -1000, 126, 57, 57, 140, 145,
	338, -1000, -1000, -1000, 788, 76, -1000, -1000, -1000, 780,
	76, 525, 525, -1000, 76, -1000, -1000, -1, -1000, 8,
	76, 27, 525, -1000, -1000, 525, -1000, 390, 141, -1000,
	42, 76, -1000, -1000, 338, -1000, -8, -37, 603, -1000,
	525, 120, 525, 166, -1000, -1000, 121, 603, 76, 338,
	-1000, 525, -1000, -1000, -38, -1000, -1000, -1000, 76, 191,
	-1, -20, 603, -1000, 235, -1000, 103, -1000, -1000, -1000,
	-1000, -1000,
}
var yyPgo = [...]int{

	0, 8, 9, 256, 255, 5, 253, 251, 0, 10,
	101, 16, 83, 250, 97, 244, 242, 15, 241, 11,
	18, 2, 235, 7, 230, 228, 224, 223, 221, 1,
	14, 220, 13, 218, 216, 6, 12, 214, 3, 212,
	209, 208, 205,
}
var yyR1 = [...]int{

	0, 40, 36, 36, 41, 41, 37, 37, 37, 23,
	23, 23, 23, 24, 24, 25, 25, 25, 27, 27,
	26, 26, 28, 28, 29, 31, 31, 30, 30, 30,
	30, 30, 30, 30, 30, 42, 42, 11, 11, 11,
	11, 11, 11, 11, 11, 11, 11, 11, 11, 11,
	11, 11, 4, 4, 3, 3, 2, 2, 2, 2,
	39, 39, 38, 38, 7, 7, 6, 6, 5, 5,
	5, 5, 5, 12, 12, 13, 13, 15, 15, 16,
	16, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8,
	14, 14, 9, 9, 10, 10, 1, 1, 32, 34,
	34, 33, 33, 33, 17, 17, 35, 21, 22, 22,
	22, 22, 18, 19, 19, 20, 20,
}
var yyR2 = [...]int{

	0, 2, 5, 2, 0, 2, 0, 3, 2, 0,
	2, 2, 3, 1, 1, 7, 6, 1, 4, 5,
	1, 4, 2, 1, 4, 0, 3, 1, 2, 1,
	3, 3, 1, 1, 1, 0, 1, 1, 1, 1,
	3, 7, 4, 4, 6, 8, 3, 4, 4, 3,
	4, 3, 0, 2, 1, 3, 1, 3, 2, 2,
	1, 3, 1, 3, 0, 2, 1, 3, 1, 3,
	2, 1, 2, 1, 3, 0, 1, 1, 3, 0,
	2, 1, 4, 2, 2, 2, 2, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	4, 3, 3, 3, 3, 3, 3, 3, 3, 5,
	1, 3, 0, 1, 0, 2, 0, 1, 3, 1,
	3, 0, 1, 2, 1, 3, 1, 1, 3, 2,
	2, 1, 4, 1, 3, 1, 2,
}
var yyChk = [...]int{

	-1000, -40, -23, 28, -24, 58, 27, -29, -25, -30,
	50, 30, -26, -12, 51, 52, 53, 54, -27, -8,
	34, -11, 39, 47, 10, 8, 24, -21, -22, -35,
	40, 17, 19, 5, 32, 33, 11, 49, 58, -31,
	32, -17, -11, 15, 25, 9, -12, -28, 35, 36,
	7, 4, 12, 44, 8, 10, 14, 16, 29, 41,
	42, 31, 37, 47, 48, 26, 21, 22, 23, 45,
	46, 38, 34, -12, 11, 5, 17, -16, -15, -12,
	-8, -8, -8, -8, 5, -10, -8, -14, -32, -33,
	-14, -8, -34, -10, -8, 11, 33, -42, 61, 5,
	37, 9, -12, -12, -8, -12, 13, 34, -8, -8,
	-8, -8, -8, -8, -8, -8, -8, -8, -8, -8,
	-8, 37, -8, -8, -8, -8, -8, -8, -8, -8,
	-8, 13, 32, -4, -3, -2, -8, -21, 7, 43,
	-12, -13, 13, 9, -1, -35, 18, -20, -19, -18,
	30, 9, -1, -20, 20, -1, 13, 9, 6, 33,
	58, -30, -7, -6, -5, -21, 7, 43, -12, -11,
	13, -36, 58, -29, -8, 35, -36, 6, -1, 9,
	15, -8, -8, 18, 13, -12, -12, 9, 18, -19,
	34, -17, -8, 20, 20, -8, -32, -8, 6, -1,
	9, 15, -21, -21, 13, -36, -37, -41, -8, -2,
	-8, -9, -8, -39, -38, -35, -21, -8, 37, 13,
	-5, -8, -36, 55, 27, 58, 58, 18, 13, -1,
	9, 15, -8, -36, -23, 58, -9, 6, -38, -35,
	56, 18,
}
var yyDef = [...]int{

	9, -2, 0, 1, 10, 11, 0, 13, 14, 25,
	0, 0, 17, 27, 29, 32, 33, 34, 20, 73,
	0, 81, 79, 0, 0, 0, 0, 37, 38, 39,
	0, 114, 121, 114, 127, 131, 0, 126, 12, 35,
	0, 0, 124, 0, 0, 0, 28, 0, 0, 23,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 52, 75, 0, 116, 77,
	83, 84, 85, 86, 0, 0, 110, 116, 119, 0,
	116, 110, 122, 0, 110, 129, 130, 0, 36, 64,
	0, 0, 30, 31, 74, 0, 0, 22, 87, 88,
	89, 90, 91, 92, 93, 94, 95, 96, 97, 98,
	99, 0, 101, 102, 103, 104, 105, 106, 107, 108,
	0, 0, 40, 0, 116, 54, 56, 37, 0, 0,
	76, 0, 0, 117, 80, 0, 46, 0, 135, 133,
	0, 117, 115, 0, 49, 0, 0, 123, 51, 128,
	24, 26, 0, 116, 66, 68, 71, 0, 0, 125,
	0, 21, 6, 4, 100, 0, 18, 42, 53, 117,
	0, 58, 59, 43, 112, 82, 78, 0, 47, 136,
	0, 0, 111, 48, 50, 118, 120, 0, 0, 65,
	117, 0, 70, 72, 0, 19, 0, 3, 109, 55,
	57, 0, 113, 116, 60, 62, 0, 134, 0, 0,
	67, 69, 16, 9, 0, 8, 5, 44, 112, 0,
	117, 0, 132, 15, 0, 7, 0, 41, 61, 63,
	2, 45,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	58, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 4, 22, 3,
	5, 6, 7, 8, 9, 10, 11, 12, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 13, 61,
	14, 15, 16, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 17, 3, 18, 23, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 19, 21, 20, 24,
}
var yyTok2 = [...]int{

	2, 3, 25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40, 41, 42,
	43, 44, 45, 46, 47, 48, 49, 50, 51, 52,
	53, 54, 55, 56, 57, 59, 60,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:195
		{
			yylex.(*input).file = &File{Stmt: yyDollar[1].exprs}
			return 0
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:202
		{
			statements := yyDollar[4].exprs
			if yyDollar[2].exprs != nil {
				// $2 can only contain *CommentBlock objects, each of them contains a non-empty After slice
				cb := yyDollar[2].exprs[len(yyDollar[2].exprs)-1].(*CommentBlock)
				// $4 can't be empty and can't start with a comment
				stmt := yyDollar[4].exprs[0]
				start, _ := stmt.Span()
				if start.Line-cb.After[len(cb.After)-1].Start.Line == 1 {
					// The first statement of $4 starts on the next line after the last comment of $2.
					// Attach the last comment to the first statement
					stmt.Comment().Before = cb.After
					yyDollar[2].exprs = yyDollar[2].exprs[:len(yyDollar[2].exprs)-1]
				}
				statements = append(yyDollar[2].exprs, yyDollar[4].exprs...)
			}
			yyVAL.exprs = statements
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 3:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:222
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 6:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:230
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 7:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:235
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].lastStmt
			if yyVAL.lastStmt == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastStmt = cb
			}
			com := yyVAL.lastStmt.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:247
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 9:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:253
		{
			yyVAL.exprs = nil
			yyVAL.lastStmt = nil
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:258
		{
			// If this statement follows a comment block,
			// attach the comments to the statement.
			if cb, ok := yyDollar[1].lastStmt.(*CommentBlock); ok {
				yyVAL.exprs = append(yyDollar[1].exprs[:len(yyDollar[1].exprs)-1], yyDollar[2].exprs...)
				yyDollar[2].exprs[0].Comment().Before = cb.After
				yyVAL.lastStmt = yyDollar[2].lastStmt
				break
			}

			// Otherwise add to list.
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyDollar[2].lastStmt

			// Consider this input:
			//
			//	foo()
			//	# bar
			//	baz()
			//
			// If we've just parsed baz(), the # bar is attached to
			// foo() as an After comment. Make it a Before comment
			// for baz() instead.
			if x := yyDollar[1].lastStmt; x != nil {
				com := x.Comment()
				// stmt is never empty
				yyDollar[2].exprs[0].Comment().Before = com.After
				com.After = nil
			}
		}
	case 11:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:289
		{
			// Blank line; sever last rule from future comments.
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = nil
		}
	case 12:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:295
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].lastStmt
			if yyVAL.lastStmt == nil {
				cb := &CommentBlock{Start: yyDollar[2].pos}
				yyVAL.exprs = append(yyVAL.exprs, cb)
				yyVAL.lastStmt = cb
			}
			com := yyVAL.lastStmt.Comment()
			com.After = append(com.After, Comment{Start: yyDollar[2].pos, Token: yyDollar[2].tok})
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:309
		{
			yyVAL.exprs = yyDollar[1].exprs
			yyVAL.lastStmt = yyDollar[1].exprs[len(yyDollar[1].exprs)-1]
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:314
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
			yyVAL.lastStmt = yyDollar[1].expr
			if cbs := extractTrailingComments(yyDollar[1].expr); len(cbs) > 0 {
				yyVAL.exprs = append(yyVAL.exprs, cbs...)
				yyVAL.lastStmt = cbs[len(cbs)-1]
				if yyDollar[1].lastStmt == nil {
					yyVAL.lastStmt = nil
				}
			}
		}
	case 15:
		yyDollar = yyS[yypt-7 : yypt+1]
//line build/parse.y:328
		{
			yyVAL.expr = &DefStmt{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[4].exprs,
					Body:     yyDollar[7].exprs,
				},
				Name:           yyDollar[2].tok,
				ColonPos:       yyDollar[6].pos,
				ForceCompact:   forceCompact(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
				ForceMultiLine: forceMultiLine(yyDollar[3].pos, yyDollar[4].exprs, yyDollar[5].pos),
			}
			yyVAL.lastStmt = yyDollar[7].lastStmt
		}
	case 16:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:343
		{
			yyVAL.expr = &ForStmt{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				X:    yyDollar[4].expr,
				Body: yyDollar[6].exprs,
			}
			yyVAL.lastStmt = yyDollar[6].lastStmt
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:353
		{
			yyVAL.expr = yyDollar[1].ifstmt
			yyVAL.lastStmt = yyDollar[1].lastStmt
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:361
		{
			yyVAL.ifstmt = &IfStmt{
				If:   yyDollar[1].pos,
				Cond: yyDollar[2].expr,
				True: yyDollar[4].exprs,
			}
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 19:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:370
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = []Expr{
				&IfStmt{
					If:   yyDollar[2].pos,
					Cond: yyDollar[3].expr,
					True: yyDollar[5].exprs,
				},
			}
			yyVAL.lastStmt = yyDollar[5].lastStmt
		}
	case 21:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:391
		{
			yyVAL.ifstmt = yyDollar[1].ifstmt
			inner := yyDollar[1].ifstmt
			for len(inner.False) == 1 {
				inner = inner.False[0].(*IfStmt)
			}
			inner.ElsePos = End{Pos: yyDollar[2].pos}
			inner.False = yyDollar[4].exprs
			yyVAL.lastStmt = yyDollar[4].lastStmt
		}
	case 24:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:408
		{
			yyVAL.exprs = append([]Expr{yyDollar[1].expr}, yyDollar[2].exprs...)
			yyVAL.lastStmt = yyVAL.exprs[len(yyVAL.exprs)-1]
		}
	case 25:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:414
		{
			yyVAL.exprs = []Expr{}
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:418
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:425
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
				Result: yyDollar[2].expr,
			}
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:432
		{
			yyVAL.expr = &ReturnStmt{
				Return: yyDollar[1].pos,
			}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:437
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:438
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:440
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 33:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:447
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 34:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:454
		{
			yyVAL.expr = &BranchStmt{
				Token:    yyDollar[1].tok,
				TokenPos: yyDollar[1].pos,
			}
		}
	case 39:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:468
		{
			yyVAL.expr = yyDollar[1].string
		}
	case 40:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:472
		{
			yyVAL.expr = &DotExpr{
				X:       yyDollar[1].expr,
				Dot:     yyDollar[2].pos,
				NamePos: yyDollar[3].pos,
				Name:    yyDollar[3].tok,
			}
		}
	case 41:
		yyDollar = yyS[yypt-7 : yypt+1]
//line build/parse.y:481
		{
			load := &LoadStmt{
				Load:         yyDollar[1].pos,
				Module:       yyDollar[3].string,
				Rparen:       End{Pos: yyDollar[7].pos},
				ForceCompact: yyDollar[1].pos.Line == yyDollar[7].pos.Line,
			}
			for _, arg := range yyDollar[5].loadargs {
				load.From = append(load.From, &arg.from)
				load.To = append(load.To, &arg.to)
			}
			yyVAL.expr = load
		}
	case 42:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:495
		{
			yyVAL.expr = &CallExpr{
				X:              yyDollar[1].expr,
				ListStart:      yyDollar[2].pos,
				List:           yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceCompact:   forceCompact(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
				ForceMultiLine: forceMultiLine(yyDollar[2].pos, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 43:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:506
		{
			yyVAL.expr = &IndexExpr{
				X:          yyDollar[1].expr,
				IndexStart: yyDollar[2].pos,
				Y:          yyDollar[3].expr,
				End:        yyDollar[4].pos,
			}
		}
	case 44:
		yyDollar = yyS[yypt-6 : yypt+1]
//line build/parse.y:515
		{
			yyVAL.expr = &SliceExpr{
				X:          yyDollar[1].expr,
				SliceStart: yyDollar[2].pos,
				From:       yyDollar[3].expr,
				FirstColon: yyDollar[4].pos,
				To:         yyDollar[5].expr,
				End:        yyDollar[6].pos,
			}
		}
	case 45:
		yyDollar = yyS[yypt-8 : yypt+1]
//line build/parse.y:526
		{
			yyVAL.expr = &SliceExpr{
				X:           yyDollar[1].expr,
				SliceStart:  yyDollar[2].pos,
				From:        yyDollar[3].expr,
				FirstColon:  yyDollar[4].pos,
				To:          yyDollar[5].expr,
				SecondColon: yyDollar[6].pos,
				Step:        yyDollar[7].expr,
				End:         yyDollar[8].pos,
			}
		}
	case 46:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:539
		{
			yyVAL.expr = &ListExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
			}
		}
	case 47:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:548
		{
			yyVAL.expr = &Comprehension{
				Curly:          false,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].expr,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLineComprehension(yyDollar[1].pos, yyDollar[2].expr, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 48:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:559
		{
			yyVAL.expr = &Comprehension{
				Curly:          true,
				Lbrack:         yyDollar[1].pos,
				Body:           yyDollar[2].kv,
				Clauses:        yyDollar[3].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLineComprehension(yyDollar[1].pos, yyDollar[2].kv, yyDollar[3].exprs, yyDollar[4].pos),
			}
		}
	case 49:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:570
		{
			exprValues := make([]Expr, 0, len(yyDollar[2].kvs))
			for _, kv := range yyDollar[2].kvs {
				exprValues = append(exprValues, Expr(kv))
			}
			yyVAL.expr = &DictExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].kvs,
				End:            End{Pos: yyDollar[3].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, exprValues, yyDollar[3].pos),
			}
		}
	case 50:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:583
		{
			yyVAL.expr = &SetExpr{
				Start:          yyDollar[1].pos,
				List:           yyDollar[2].exprs,
				End:            End{Pos: yyDollar[4].pos},
				ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[4].pos),
			}
		}
	case 51:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:592
		{
			if len(yyDollar[2].exprs) == 1 && yyDollar[2].comma.Line == 0 {
				// Just a parenthesized expression, not a tuple.
				yyVAL.expr = &ParenExpr{
					Start:          yyDollar[1].pos,
					X:              yyDollar[2].exprs[0],
					End:            End{Pos: yyDollar[3].pos},
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			} else {
				yyVAL.expr = &TupleExpr{
					Start:          yyDollar[1].pos,
					List:           yyDollar[2].exprs,
					End:            End{Pos: yyDollar[3].pos},
					ForceCompact:   forceCompact(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
					ForceMultiLine: forceMultiLine(yyDollar[1].pos, yyDollar[2].exprs, yyDollar[3].pos),
				}
			}
		}
	case 52:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:613
		{
			yyVAL.exprs = nil
		}
	case 53:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:617
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 54:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:623
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 55:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:627
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 57:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:634
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 58:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:638
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 59:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:642
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 60:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:647
		{
			yyVAL.loadargs = []*struct {
				from Ident
				to   Ident
			}{yyDollar[1].loadarg}
		}
	case 61:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:651
		{
			yyDollar[1].loadargs = append(yyDollar[1].loadargs, yyDollar[3].loadarg)
			yyVAL.loadargs = yyDollar[1].loadargs
		}
	case 62:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:657
		{
			start := yyDollar[1].string.Start.add("'")
			if yyDollar[1].string.TripleQuote {
				start = start.add("''")
			}
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: start,
				},
				to: Ident{
					Name:    yyDollar[1].string.Value,
					NamePos: start,
				},
			}
		}
	case 63:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:674
		{
			start := yyDollar[3].string.Start.add("'")
			if yyDollar[3].string.TripleQuote {
				start = start.add("''")
			}
			yyVAL.loadarg = &struct {
				from Ident
				to   Ident
			}{
				from: Ident{
					Name:    yyDollar[3].string.Value,
					NamePos: start,
				},
				to: *yyDollar[1].expr.(*Ident),
			}
		}
	case 64:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:689
		{
			yyVAL.exprs = nil
		}
	case 65:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:693
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 66:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:699
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 67:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:703
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 69:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:710
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 70:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:714
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 71:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:718
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, nil)
		}
	case 72:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:722
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 74:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:729
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.NoBrackets {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					NoBrackets:     true,
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 75:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:744
		{
			yyVAL.expr = nil
		}
	case 77:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:751
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 78:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:755
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 79:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:760
		{
			yyVAL.exprs = nil
		}
	case 80:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:764
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 82:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:771
		{
			yyVAL.expr = &LambdaExpr{
				Function: Function{
					StartPos: yyDollar[1].pos,
					Params:   yyDollar[2].exprs,
					Body:     []Expr{yyDollar[4].expr},
				},
			}
		}
	case 83:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:780
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 84:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:781
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 85:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:782
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 86:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:783
		{
			yyVAL.expr = unary(yyDollar[1].pos, yyDollar[1].tok, yyDollar[2].expr)
		}
	case 87:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:784
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 88:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:785
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 89:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:786
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 90:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:787
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 91:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:788
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 92:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:789
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 93:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:790
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 94:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:791
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 95:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:792
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 96:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:793
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 97:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:794
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 98:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:795
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 99:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:796
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 100:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:797
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "not in", yyDollar[4].expr)
		}
	case 101:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:798
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 102:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:799
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 103:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:800
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 104:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:801
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 105:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:802
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 106:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:803
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 107:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:804
		{
			yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
		}
	case 108:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:806
		{
			if b, ok := yyDollar[3].expr.(*UnaryExpr); ok && b.Op == "not" {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, "is not", b.X)
			} else {
				yyVAL.expr = binary(yyDollar[1].expr, yyDollar[2].pos, yyDollar[2].tok, yyDollar[3].expr)
			}
		}
	case 109:
		yyDollar = yyS[yypt-5 : yypt+1]
//line build/parse.y:814
		{
			yyVAL.expr = &ConditionalExpr{
				Then:      yyDollar[1].expr,
				IfStart:   yyDollar[2].pos,
				Test:      yyDollar[3].expr,
				ElseStart: yyDollar[4].pos,
				Else:      yyDollar[5].expr,
			}
		}
	case 110:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:826
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 111:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:830
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[3].expr)
		}
	case 112:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:835
		{
			yyVAL.expr = nil
		}
	case 114:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:841
		{
			yyVAL.exprs, yyVAL.comma = nil, Position{}
		}
	case 115:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:845
		{
			yyVAL.exprs, yyVAL.comma = yyDollar[1].exprs, yyDollar[2].pos
		}
	case 116:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:854
		{
			yyVAL.pos = Position{}
		}
	case 118:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:860
		{
			yyVAL.kv = &KeyValueExpr{
				Key:   yyDollar[1].expr,
				Colon: yyDollar[2].pos,
				Value: yyDollar[3].expr,
			}
		}
	case 119:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:870
		{
			yyVAL.kvs = []*KeyValueExpr{yyDollar[1].kv}
		}
	case 120:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:874
		{
			yyVAL.kvs = append(yyDollar[1].kvs, yyDollar[3].kv)
		}
	case 121:
		yyDollar = yyS[yypt-0 : yypt+1]
//line build/parse.y:879
		{
			yyVAL.kvs = nil
		}
	case 122:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:883
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 123:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:887
		{
			yyVAL.kvs = yyDollar[1].kvs
		}
	case 125:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:894
		{
			tuple, ok := yyDollar[1].expr.(*TupleExpr)
			if !ok || !tuple.NoBrackets {
				tuple = &TupleExpr{
					List:           []Expr{yyDollar[1].expr},
					NoBrackets:     true,
					ForceCompact:   true,
					ForceMultiLine: false,
				}
			}
			tuple.List = append(tuple.List, yyDollar[3].expr)
			yyVAL.expr = tuple
		}
	case 126:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:910
		{
			yyVAL.string = &StringExpr{
				Start:       yyDollar[1].pos,
				Value:       yyDollar[1].str,
				TripleQuote: yyDollar[1].triple,
				End:         yyDollar[1].pos.add(yyDollar[1].tok),
				Token:       yyDollar[1].tok,
			}
		}
	case 127:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:922
		{
			yyVAL.expr = &Ident{NamePos: yyDollar[1].pos, Name: yyDollar[1].tok}
		}
	case 128:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:928
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "." + yyDollar[3].tok}
		}
	case 129:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:932
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok + "."}
		}
	case 130:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:936
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: "." + yyDollar[2].tok}
		}
	case 131:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:940
		{
			yyVAL.expr = &LiteralExpr{Start: yyDollar[1].pos, Token: yyDollar[1].tok}
		}
	case 132:
		yyDollar = yyS[yypt-4 : yypt+1]
//line build/parse.y:946
		{
			yyVAL.expr = &ForClause{
				For:  yyDollar[1].pos,
				Vars: yyDollar[2].expr,
				In:   yyDollar[3].pos,
				X:    yyDollar[4].expr,
			}
		}
	case 133:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:956
		{
			yyVAL.exprs = []Expr{yyDollar[1].expr}
		}
	case 134:
		yyDollar = yyS[yypt-3 : yypt+1]
//line build/parse.y:959
		{
			yyVAL.exprs = append(yyDollar[1].exprs, &IfClause{
				If:   yyDollar[2].pos,
				Cond: yyDollar[3].expr,
			})
		}
	case 135:
		yyDollar = yyS[yypt-1 : yypt+1]
//line build/parse.y:968
		{
			yyVAL.exprs = yyDollar[1].exprs
		}
	case 136:
		yyDollar = yyS[yypt-2 : yypt+1]
//line build/parse.y:971
		{
			yyVAL.exprs = append(yyDollar[1].exprs, yyDollar[2].exprs...)
		}
	}
	goto yystack /* stack new state and value */
}
