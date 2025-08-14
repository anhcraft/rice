package frontend

import (
	"errors"
	"rice/exec/ast"
	"rice/exec/ast/opr"
)

// TODO update AST position tracking to correctly identify problem spots

var (
	limIfDepth                       = 64
	limCallArgsBreadth               = 64
	limCallParamsBreadth             = 64 // should equal to limCallArgsBreadth
	expectIdErr                      = errors.New("expect identifier")
	expectReferenceErr               = errors.New("expect identifier, element access or selector")
	expectValueErr                   = errors.New("expect identifier, literal or expression")
	expectValueLparenForErr          = errors.New("expect '(', identifier, literal or expression")
	expectPartSeparatorErr           = errors.New("expect ';' to separate parts")
	expectValuePartSeparatorErr      = errors.New("expect ';', identifier, literal or expression")
	expectSimpleStmtPartSeparatorErr = errors.New("expect ';', declaration, increment, decrement, identifier, literal or expression")
	expectRbracketElemAccessErr      = errors.New("expect ']' to finish element-access expression")
	expectRparenGroupErr             = errors.New("expect ')' to finish group expression")
	expectRparenForClauseErr         = errors.New("expect ')' to finish for clause")
	expectSimpleStmtRparenErr        = errors.New("expect ')', declaration, increment, decrement, identifier, literal or expression")
	expectValueRparenCallArgsErr     = errors.New("expect identifier, literal, expression or ')' to finish call argument list")
	expectCommaRparenCallArgsErr     = errors.New("expect ',' to continue or ')' to finish call argument list")
	expectIdRparenFuncParamsErr      = errors.New("expect identifier or ')' to finish function parameter list")
	expectCommaRparenFuncParamsErr   = errors.New("expect ',' to continue or ')' to finish function parameter list")
	expectLbraceBlockErr             = errors.New("expect '{' before block expression")
	expectRbraceBlockErr             = errors.New("expect '}' to finish block expression")
	expectLbraceOrIfBranchErr        = errors.New("expect '{' or 'if' after 'else' branch")
	expectLparenFuncErr              = errors.New("expect '(' to specify parameter list")
	expectEqualRhsErr                = errors.New("expect '=' to begin the right-hand side")
	expectStmtTerminatorErr          = errors.New("expect statement terminator: '}', ';' or eof")
	genericUnexpectedTokenErr        = errors.New("unexpected token")
	excessiveIfDepthErr              = errors.New("'if' block is nested too depth")
	excessiveCallArgsSizeErr         = errors.New("too many call arguments")
	excessiveFuncParamsSizeErr       = errors.New("too many function parameters")
	invalidAssignmentTargetErr       = errors.New("invalid assignment target")
)

type Parser struct {
	tokens []Token
	cursor int   // the current token pos
	prior  Token // the prior token

	panicMode bool          // on panicMode, skip tokens until the recovery-point
	errors    []SyntaxError // all errors collected
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) eof() bool {
	return p.cursor >= len(p.tokens) || p.tokens[p.cursor].tokenType == eof
}

func (p *Parser) next() {
	if !p.eof() {
		p.prior = p.tokens[p.cursor]
		p.cursor++
	}
}

func (p *Parser) peek() Token {
	if p.eof() {
		return Token{tokenType: eof}
	}
	return p.tokens[p.cursor]
}

// captureCurrStart takes the starting Pos of the current token
func (p *Parser) captureCurrStart() Pos {
	if p.cursor >= len(p.tokens) {
		panic("out of bound")
	}
	return p.tokens[p.cursor].start
}

// captureLastEnd takes the ending Pos of the prior token
func (p *Parser) capturePriorEnd() Pos {
	if p.cursor == 0 {
		panic("out of bound")
	}
	return p.prior.end
}

// match checks if the current token is one of the given types
func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.peek().tokenType == t {
			return true
		}
	}
	return false
}

// throw reports an error and enters panic mode
func (p *Parser) throw(err error, start Pos) {
	if p.panicMode {
		return
	}
	p.panicMode = true

	p.errors = append(p.errors, SyntaxError{
		Err:   err,
		Start: start,
		End:   p.capturePriorEnd(),
	})
}

// [start]--- SINGLE STATEMENT ---

func (p *Parser) takeStmt() ast.Stmt {
	off := p.captureCurrStart()

	if p.match(returnKeyword) {
		p.next()

		var val ast.Expr

		if !p.match(eof, semicolon, rbrace) {
			val = p.takeExpr()
		}

		return &ast.ControlStmt{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			OpPos: off.ast(),
			Op:    opr.Return,
			Value: val,
		}
	} else if p.match(breakKeyword) {
		p.next()

		return &ast.ControlStmt{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			OpPos: off.ast(),
			Op:    opr.Break,
		}
	} else if p.match(continueKeyword) {
		p.next()

		return &ast.ControlStmt{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			OpPos: off.ast(),
			Op:    opr.Continue,
		}
	} else if p.match(forKeyword) {
		return p.takeForStmt()
	} else {
		return p.takeSimpleStmt()
	}
}

func (p *Parser) takeSimpleStmt() ast.SimpleStmt {
	var stmt ast.SimpleStmt

	if p.match(varKeyword, constKeyword) {
		stmt = p.takeDeclStmt()
	} else {
		// handle X++, ++X, X--, --X
		// or fallback to expression
		stmt = p.takeIncDecStmtOrExpr()
	}

	return stmt
}

func (p *Parser) takeDeclStmt() ast.SimpleStmt {
	off := p.captureCurrStart()
	constant := false

	if p.match(varKeyword) {
		constant = false
	} else if p.match(constKeyword) {
		constant = true
	} else {
		panic("expected var or const")
	}
	p.next()

	var target *ast.IdentifierExpr
	left := p.takePrimary()

	if left == ast.Invalid {
		p.throw(expectIdErr, off)
		return ast.Invalid
	}

	switch t := left.(type) {
	case *ast.IdentifierExpr:
		target = t
	default:
		p.throw(expectIdErr, off)
		return ast.Invalid
	}

	if p.match(equal) {
		p.next()

		value := p.takeRightAssocBinary()

		if value == ast.Invalid {
			p.throw(expectValueErr, off)
			return ast.Invalid
		}

		return &ast.DeclareStmt{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Const:  constant,
			Target: target,
			Value:  value,
		}
	}

	p.throw(expectEqualRhsErr, off)
	return ast.Invalid
}

func (p *Parser) takeIncDecStmtOrExpr() ast.SimpleStmt {
	off := p.captureCurrStart()
	pre := false
	var target ast.Expr
	var op opr.OpType
	var opPos Pos

	// prefix
	if p.match(increment, decrement) {
		if p.match(increment) {
			op = opr.Inc
		} else {
			op = opr.Dec
		}
		opPos = p.captureCurrStart()
		p.next()

		pre = true
		target = p.takePostUnary()
		if !takeIncDecStmtOrExprValidator(target) {
			p.throw(expectReferenceErr, off)
			return ast.Invalid
		}
	} else { // suffix
		target = p.takeExpr()

		if takeIncDecStmtOrExprValidator(target) && p.match(increment, decrement) {
			if p.match(increment) {
				op = opr.Inc
			} else {
				op = opr.Dec
			}
			opPos = p.captureCurrStart()
			p.next()
		} else {
			return target
		}
	}

	return &ast.IncDecStmt{
		BaseNode: ast.BaseNode{
			Start: off.ast(),
			End:   p.captureCurrStart().ast(),
		},
		Pre:    pre,
		Op:     op,
		OpPos:  opPos.ast(),
		Target: target,
	}
}

func takeIncDecStmtOrExprValidator(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.IdentifierExpr:
		return true
	case *ast.ElementAccessExpr:
		return true
	case *ast.SelectorExpr:
		return true
	}
	return false
}

func (p *Parser) takeForStmt() ast.Stmt {
	off := p.captureCurrStart()
	if !p.match(forKeyword) {
		panic("expected for")
	}
	p.next()

	if p.match(lparen) { // C-style
		p.next()

		var init ast.SimpleStmt

		if p.match(semicolon) {
			p.next()
		} else {
			init = p.takeSimpleStmt()

			if init == ast.Invalid {
				p.throw(expectSimpleStmtPartSeparatorErr, off)
				return ast.Invalid
			}

			if p.match(semicolon) {
				p.next()
			} else {
				p.throw(expectPartSeparatorErr, off)
				return ast.Invalid
			}
		}

		var cond ast.Expr

		if p.match(semicolon) {
			p.next()
		} else {
			cond = p.takeExpr()

			if cond == ast.Invalid {
				p.throw(expectValuePartSeparatorErr, off)
				return ast.Invalid
			}

			if p.match(semicolon) {
				p.next()
			} else {
				p.throw(expectPartSeparatorErr, off)
				return ast.Invalid
			}
		}

		var post ast.SimpleStmt

		if p.match(rparen) {
			p.next()
		} else {
			post = p.takeSimpleStmt()

			if post == ast.Invalid {
				p.throw(expectSimpleStmtRparenErr, off)
				return ast.Invalid
			}

			if p.match(rparen) {
				p.next()
			} else {
				p.throw(expectRparenForClauseErr, off)
				return ast.Invalid
			}
		}

		var body *ast.BlockExpr

		if p.match(lbrace) {
			body = p.takePrimaryBlockExpr()
		} else {
			p.throw(expectLbraceBlockErr, off)
			return ast.Invalid
		}

		return &ast.ForStmt{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Init: init,
			Cond: cond,
			Post: post,
			Body: body,
		}
	} else { // Short-form / For-in
		first := p.takeExpr()
		if first == ast.Invalid {
			p.throw(expectValueLparenForErr, off)
			return nil
		}

		if key, ok := first.(*ast.IdentifierExpr); ok {
			if p.match(inKeyword) {
				p.next()

				val := p.takeExpr()
				if val == ast.Invalid {
					p.throw(expectValueErr, off)
					return nil
				}

				var body *ast.BlockExpr

				if p.match(lbrace) {
					body = p.takePrimaryBlockExpr()
				} else {
					p.throw(expectLbraceBlockErr, off)
					return ast.Invalid
				}

				return &ast.ForInStmt{
					BaseNode: ast.BaseNode{
						Start: off.ast(),
						End:   p.captureCurrStart().ast(),
					},
					Key:   key,
					Value: val,
					Body:  body,
				}
			}
		}

		var body *ast.BlockExpr

		if p.match(lbrace) {
			body = p.takePrimaryBlockExpr()
		} else {
			p.throw(expectLbraceBlockErr, off)
			return ast.Invalid
		}

		return &ast.ForStmt{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Init: nil,
			Cond: first,
			Post: nil,
			Body: body,
		}
	}
}

// --- SINGLE STATEMENT ---[end]

// -------------------------------

// [start]--- SINGLE EXPRESSION ---

// takeExpr takes a single expression
func (p *Parser) takeExpr() ast.Expr {
	return p.takeRightAssocBinary()
}

// takeRightAssocBinary (lowest precedence)
func (p *Parser) takeRightAssocBinary() ast.Expr {
	off := p.captureCurrStart()
	expr := p.takeLeftAssocBinary()

	if p.match(equal) {
		p.next()

		value := p.takeLeftAssocBinary()
		if value == ast.Invalid {
			p.throw(expectValueErr, off)
			return nil
		}

		var target ast.Expr

		switch t := expr.(type) {
		case *ast.IdentifierExpr:
			target = t
		case *ast.ElementAccessExpr:
			target = t
		default:
			p.throw(invalidAssignmentTargetErr, off)
			return ast.Invalid
		}

		return &ast.AssignExpr{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Target: target,
			Value:  value,
		}
	}

	return expr
}

// takeLeftAssocBinary
func (p *Parser) takeLeftAssocBinary() ast.Expr {
	// from the most precedence to the least
	takeMulti := func() ast.Expr {
		return p.takeLeftAssocBinaryHelper(p.takePreUnary, true, star, slash, percent)
	}
	takeAdd := func() ast.Expr {
		return p.takeLeftAssocBinaryHelper(takeMulti, true, plus, minus)
	}
	takeCmpRe := func() ast.Expr {
		return p.takeLeftAssocBinaryHelper(takeAdd, false, gt, gte, lt, lte)
	}
	takeCmpEq := func() ast.Expr {
		return p.takeLeftAssocBinaryHelper(takeCmpRe, false, equalEqual, bangEqual)
	}
	takeAnd := func() ast.Expr {
		return p.takeLeftAssocBinaryHelper(takeCmpEq, true, and)
	}
	takeOr := func() ast.Expr {
		return p.takeLeftAssocBinaryHelper(takeAnd, true, or)
	}

	return takeOr()
}

// takeLeftAssocBinaryHelper handles binary operator and cascades to the next precedence
func (p *Parser) takeLeftAssocBinaryHelper(nextPrecedence func() ast.Expr, chainable bool, types ...TokenType) ast.Expr {
	off := p.captureCurrStart()
	expr := nextPrecedence()

	for p.match(types...) {
		op := p.peek()
		p.next()
		right := nextPrecedence()
		if right == ast.Invalid {
			p.throw(expectValueErr, off)
			return nil
		}

		var opt opr.OpType

		switch op.tokenType {
		case star:
			opt = opr.Multi
		case percent:
			opt = opr.Rem
		case slash:
			opt = opr.Div
		case plus:
			opt = opr.Sum
		case minus:
			opt = opr.Sub
		case equalEqual:
			opt = opr.Eq
		case bangEqual:
			opt = opr.Neq
		case and:
			opt = opr.And
		case or:
			opt = opr.Or
		case gt:
			opt = opr.Gt
		case gte:
			opt = opr.Gte
		case lt:
			opt = opr.Lt
		case lte:
			opt = opr.Lte
		default:
			panic("unsupported")
		}

		expr = &ast.BinaryExpr{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Left:  expr,
			Op:    opt,
			OpPos: op.start.ast(),
			Right: right,
		}

		if !chainable {
			return expr
		}
	}

	return expr
}

// takePreUnary handles prefix unary with right-to-left associativity
func (p *Parser) takePreUnary() ast.Expr {
	// strategy: recursion
	off := p.captureCurrStart()

	if p.match(minus, bang) {
		op := p.peek()
		p.next()

		right := p.takePreUnary()
		if right == ast.Invalid {
			p.throw(expectValueErr, off)
			return nil
		}

		var opt opr.OpType

		switch op.tokenType {
		case minus:
			opt = opr.Neg
		case bang:
			opt = opr.Inv
		default:
			panic("unsupported")
		}

		return &ast.UnaryExpr{
			OpPos: op.start.ast(),
			Op:    opt,
			Right: right,
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
		}
	}

	return p.takePostUnary()
}

// takePostUnary handles postfix unary with left-to-right associativity
func (p *Parser) takePostUnary() ast.Expr {
	// strategy: iteration

	off := p.captureCurrStart()
	expr := p.takePrimary()

	for {
		if p.match(lparen) { // Call arguments
			args := p.takePostUnaryCallArgs()
			if args == nil {
				return ast.Invalid
			}
			expr = &ast.CallExpr{
				BaseNode: ast.BaseNode{
					Start: off.ast(),
					End:   p.captureCurrStart().ast(),
				},
				Callee:    expr,
				Arguments: args,
			}
		} else if p.match(lbracket) { // Element access
			p.next()

			seq := p.takeExpr()

			if seq == ast.Invalid {
				p.throw(expectValueErr, off)
				return nil
			}

			if p.match(rbracket) {
				p.next()
			} else {
				p.throw(expectRbracketElemAccessErr, off)
				return ast.Invalid
			}

			expr = &ast.ElementAccessExpr{
				BaseNode: ast.BaseNode{
					Start: off.ast(),
					End:   p.captureCurrStart().ast(),
				},
				Object: expr,
				Index:  seq,
			}
		} else if p.match(dot) { // Selector
			p.next()

			target := p.takePrimary()

			if v, ok := target.(*ast.IdentifierExpr); ok {
				expr = &ast.SelectorExpr{
					BaseNode: ast.BaseNode{
						Start: off.ast(),
						End:   p.captureCurrStart().ast(),
					},
					Object: expr,
					Target: v,
				}
			} else {
				p.throw(expectIdErr, off)
				return ast.Invalid
			}

		} else {
			break
		}
	}

	return expr
}

// takePostUnaryCallArgs takes the next `(...)` argument list
func (p *Parser) takePostUnaryCallArgs() []ast.CallExprArg {
	if !p.match(lparen) {
		panic("expected lparen")
	}
	off := p.captureCurrStart()
	p.next()

	//goland:noinspection GoPreferNilSlice
	args := []ast.CallExprArg{}
	count := 0

	for !p.match(rparen) {
		if count >= limCallArgsBreadth {
			p.throw(excessiveCallArgsSizeErr, off)
			return nil
		}
		count++

		arg := ast.CallExprArg{}

		if p.match(ellipsis) {
			p.next()
			arg.Spread = true
		}

		val := p.takeExpr()

		if val == ast.Invalid {
			p.throw(expectValueRparenCallArgsErr, off)
			return nil
		}

		arg.Value = val
		args = append(args, arg)

		if p.match(comma) {
			// allow the last parameter to have an optional trailing comma
			p.next()
		} else {
			break // handle later
		}
	}

	if p.match(rparen) { // recheck again in case EOF
		p.next()
	} else {
		p.throw(expectCommaRparenCallArgsErr, off)
		return nil
	}

	return args
}

// takePrimary (highest precedence)
func (p *Parser) takePrimary() ast.Expr {
	off := p.captureCurrStart()

	// Literals
	if p.match(integerLiteral, floatLiteral, stringLiteral, booleanLiteral, nullLiteral) {
		tok := p.peek()
		p.next()
		return &ast.LiteralExpr{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Value: tok.literal,
		}
	}

	// Identifier (function names)
	if p.match(identifier) {
		tok := p.peek()
		p.next()
		return &ast.IdentifierExpr{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Value: tok.literal.(string),
		}
	}

	// Grouped expression
	if p.match(lparen) {
		p.next()
		expr := p.takeExpr()
		if p.match(rparen) {
			p.next()
			return expr
		}
		p.throw(expectRparenGroupErr, off)
		return ast.Invalid
	}

	// Block expression
	if p.match(lbrace) {
		block := p.takePrimaryBlockExpr()
		if block == nil {
			return ast.Invalid
		}
		return block
	}

	// If-expression
	if p.match(ifKeyword) {
		expr := p.takePrimaryIfExpr(off, 0)
		if expr == nil {
			return ast.Invalid
		}
		return expr
	}

	// Function literal
	if p.match(funcKeyword) {
		expr := p.takePrimaryFuncLitExpr(off)
		if expr == nil {
			return ast.Invalid
		}
		return expr
	}

	return ast.Invalid
}

// takePrimaryIfExpr takes the next `if` expression within an `if` subtree
func (p *Parser) takePrimaryIfExpr(rootOff Pos, depth int) *ast.IfExpr {
	if depth >= limIfDepth {
		p.throw(excessiveIfDepthErr, rootOff)
		return nil
	}

	if !p.match(ifKeyword) {
		panic("expected ifKeyword")
	}
	p.next()

	off := p.captureCurrStart()
	condition := p.takeExpr()

	if condition == ast.Invalid {
		p.throw(expectValueErr, off)
		return nil
	}

	var thenBranch *ast.BlockExpr

	if p.match(lbrace) {
		thenBranch = p.takePrimaryBlockExpr()
		if thenBranch == nil {
			return nil
		}
	} else {
		p.throw(expectLbraceBlockErr, off)
		return nil
	}

	var elseBranch ast.Expr

	if p.match(elseKeyword) {
		p.next()

		if p.match(ifKeyword) {
			elseBranch = p.takePrimaryIfExpr(rootOff, depth+1)
			if elseBranch == nil {
				return nil
			}
		} else if p.match(lbrace) {
			elseBranch = p.takePrimaryBlockExpr()
			if elseBranch == nil {
				return nil
			}
		} else {
			p.throw(expectLbraceOrIfBranchErr, off)
			return nil
		}
	}

	return &ast.IfExpr{
		BaseNode: ast.BaseNode{
			Start: off.ast(),
			End:   p.captureCurrStart().ast(),
		},
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

// takePrimaryBlockExpr takes the next `{block}` expression
func (p *Parser) takePrimaryBlockExpr() *ast.BlockExpr {
	if !p.match(lbrace) {
		panic("expected lbrace")
	}
	off := p.captureCurrStart()
	p.next() // consume lbrace

	//goland:noinspection GoPreferNilSlice
	block := []ast.Stmt{}

	if !p.match(rbrace) {
		block = p.TakeMultiStatementExpr(rbrace)
		if block == nil {
			return nil
		}
	}

	if p.match(rbrace) {
		p.next()
		return &ast.BlockExpr{
			BaseNode: ast.BaseNode{
				Start: off.ast(),
				End:   p.captureCurrStart().ast(),
			},
			Statements: block,
		}
	}
	p.throw(expectRbraceBlockErr, off)
	return &ast.BlockExpr{}
}

func (p *Parser) takePrimaryFuncLitExpr(off Pos) ast.Expr {
	if !p.match(funcKeyword) {
		panic("expected funcKeyword")
	}
	p.next()

	// take parameter list
	if !p.match(lparen) {
		p.throw(expectLparenFuncErr, off)
		return nil
	}
	params, variadic := p.takePrimaryParamList() // handle (...)
	if params == nil {
		return nil
	}

	// take function body
	if !p.match(lbrace) {
		p.throw(expectLbraceBlockErr, off)
		return nil
	}
	body := p.takePrimaryBlockExpr()
	if body == nil {
		return nil
	}

	return &ast.FuncLiteralExpr{
		BaseNode: ast.BaseNode{
			Start: off.ast(),
			End:   p.captureCurrStart().ast(),
		},
		Params:   params,
		Body:     body,
		Variadic: variadic,
	}
}

func (p *Parser) takePrimaryParamList() ([]*ast.IdentifierExpr, bool) {
	if !p.match(lparen) {
		panic("expected lparen")
	}
	off := p.captureCurrStart()
	p.next()

	variadic := false
	//goland:noinspection GoPreferNilSlice
	args := []*ast.IdentifierExpr{}
	count := 0

	for !p.match(rparen) {
		if count >= limCallParamsBreadth {
			p.throw(excessiveFuncParamsSizeErr, off)
			return nil, variadic
		}
		count++

		if p.match(identifier) {
			pr := p.takePrimary()
			args = append(args, pr.(*ast.IdentifierExpr))
		} else {
			p.throw(expectIdRparenFuncParamsErr, off)
			return nil, variadic
		}

		if p.match(comma) {
			// allow the last parameter to have an optional trailing comma
			p.next()
		} else if p.match(ellipsis) {
			p.next()
			variadic = true
			break
		} else {
			break // handle later
		}
	}

	if p.match(rparen) { // recheck again in case EOF
		p.next()
	} else {
		p.throw(expectCommaRparenFuncParamsErr, off)
		return nil, variadic
	}

	return args, variadic
}

// --- SINGLE EXPRESSION ---[end]

// -------------------------------

// [start]--- MULTI STATEMENT ---

// TakeMultiStatementExpr handles multiple sequences of statement (without braces)
func (p *Parser) TakeMultiStatementExpr(endAt TokenType) []ast.Stmt {
	//goland:noinspection GoPreferNilSlice
	arr := []ast.Stmt{}

	for !p.eof() {
		if p.match(endAt) {
			break
		}

		off := p.captureCurrStart()
		stmt := p.takeStmt()

		if stmt == ast.Invalid {
			p.synchronize()
			continue
		}

		arr = append(arr, stmt)

		_, isBlockLike := stmt.(ast.BlockLike)

		if p.match(semicolon) {
			p.next()
		} else if p.match(endAt) {
			break
		} else if isBlockLike {
			continue
		} else if !p.eof() {
			p.throw(expectStmtTerminatorErr, off)
			break
		}
	}

	return arr
}

func (p *Parser) synchronize() {
	if !p.panicMode {
		return
	}
	p.panicMode = false

	for !p.eof() {
		p.next()

		if p.prior.tokenType == semicolon {
			return
		}
	}
}

// --- MULTI STATEMENT ---[end]

// Parse parses the input and return the statement list
// return `nil` if any error occurred
func (p *Parser) Parse() []ast.Stmt {
	off := p.captureCurrStart()
	arr := p.TakeMultiStatementExpr(eof)
	if !p.eof() {
		p.throw(genericUnexpectedTokenErr, off)
		return nil
	}
	return arr
}

func (p *Parser) Errors() []SyntaxError {
	return p.errors
}
