package mss

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"strconv"

	"github.com/omniscale/magnacarto/color"
)

// Decoder decodes one or more MSS files. Parse/ParseFile can be called
// multiple times to decode dependent .mss files. MSS() returns the current
// decoded style.
type Decoder struct {
	mss           *MSS
	vars          *Properties
	scanner       *scanner
	nextTok       *token
	lastTok       *token
	expr          *expression
	lastValue     Value
	warnings      []warning
	filename      string // for warnings/errors only
	filesParsed   int
	propertyIndex int
}

type warning struct {
	line, col int
	file      string
	msg       string
}

type position struct {
	line     int
	column   int
	filename string
	filenum  int
	index    int
}

func (a position) Less(b position) bool {
	if a.filenum != b.filenum {
		return a.filenum < b.filenum
	}
	return a.line < b.line
}

func (w *warning) String() string {
	file := w.file
	if file == "" {
		file = "?"
	}
	return fmt.Sprintf("%s in %s line: %d col: %d", w.msg, file, w.line, w.col)
}

// New will allocate a new MSS Decoder
func New() *Decoder {
	mss := newMSS()
	return &Decoder{mss: mss, vars: &Properties{}, expr: &expression{}}
}

// MSS returns the current decoded style.
func (d *Decoder) MSS() *MSS {
	return d.mss
}

func (d *Decoder) next() *token {
	if d.nextTok != nil {
		tok := d.nextTok
		d.nextTok = nil
		d.lastTok = tok
		return tok
	}
	for {
		tok := d.scanner.Next()
		if tok.t == tokenError {
			d.error(d.pos(tok), tok.value)
		}
		if tok.t != tokenS && tok.t != tokenComment {
			d.lastTok = tok
			return tok
		}
	}
}

func (d *Decoder) backup() {
	if d.nextTok != nil || d.lastTok == nil {
		d.error(d.pos(d.nextTok), "internal parser bug: double backup (%v, %v)", d.nextTok, d.lastTok)
	}
	d.nextTok = d.lastTok
}

// ParseFile parses the given .mss file.
// Can be called multiple times to parse a style split into multiple files.
func (d *Decoder) ParseFile(filename string) error {
	d.filename = filename
	defer func() { d.filename = "" }()

	r, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer r.Close()
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return d.ParseString(string(content))
}

func (d *Decoder) ParseString(content string) (err error) {
	d.filesParsed += 1
	d.scanner = newScanner(content)

	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unexpected error: %v, %T", r, r)
			}
		}
	}()
	for {
		tok := d.next()
		if tok.t == tokenEOF {
			break
		}
		if tok.t == tokenError {
			return fmt.Errorf(tok.String())
		}

		d.topLevel(tok)
	}
	return err
}

// Evaluate evaluates all expressions and resolves all references to variables.
// Must be called after last ParseFile/ParseString call.
func (d *Decoder) Evaluate() (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unexpected error: %v, %T", r, r)
			}
		}
	}()

	d.evaluateProperties(d.vars, false)
	d.evaluateProperties(d.mss.Map(), true)
	for _, b := range d.mss.root.blocks {
		d.evaluateBlock(b)
	}
	return err
}

func (d *Decoder) evaluateBlock(b *block) {
	d.evaluateProperties(b.properties, true)
	for _, b := range b.blocks {
		d.evaluateBlock(b)
	}
}

func (d *Decoder) evaluateExpression(expr *expression) Value {
	// resolve all vars in the expression before evaluating it
	for i := range expr.code {
		if expr.code[i].T == typeVar {
			varname := expr.code[i].Value.(string)
			v, _ := d.vars.get(varname)
			if v == nil {
				d.error(expr.pos, "missing var %s in expression", varname)
			}
			if expr, ok := v.(*expression); ok {
				// evaluate recursive
				v = d.evaluateExpression(expr)
				d.vars.set(varname, v)
			}
			t := d.valueType(v)
			if t == typeUnknown {
				d.error(expr.pos, "unable to determine type of var %s (%v)", varname, v)
			}
			expr.code[i] = code{Value: v, T: t}
		}
	}
	v, err := expr.evaluate()
	if err != nil {
		d.error(expr.pos, "expression error: %v", err)
	}
	return v
}

func (d *Decoder) valueType(v interface{}) codeType {
	switch v.(type) {
	case string:
		return typeString
	case float64:
		return typeNum
	case color.Color:
		return typeColor
	case bool:
		return typeBool
	case []Value:
		return typeList // TODO convert v to typeList?
	default:
		return typeUnknown
	}
}

func (d *Decoder) evaluateProperties(properties *Properties, validate bool) {
	if properties == nil {
		return
	}
	for _, k := range properties.keys() {
		if expr, ok := properties.getKey(k).(*expression); ok {
			v := d.evaluateExpression(expr)
			if validate && !validProperty(k.name, v) {
				d.warn(properties.pos(k), "invalid property %v %v", k.name, v)
			}
			attr := properties.values[k]
			properties.setPos(k, v, attr.pos)
		}
	}
}

func (d *Decoder) topLevel(tok *token) {
	switch tok.t {
	case tokenAtKeyword:
		keyword := tok.value[1:]
		d.expect(tokenColon)
		d.expressionList()
		d.expect(tokenSemicolon)
		d.vars.set(keyword, d.lastValue)
	case tokenHash, tokenAttachment, tokenClass, tokenLBracket:
		d.rule(tok)
	case tokenIdent:
		if tok.value != "Map" {
			d.error(d.pos(tok), "only 'Map' identifier expected at top level, got %v", tok)
		}
		d.mss.pushMapBlock()
		d.expect(tokenLBrace)
		d.block()
		d.mss.popBlock()
		// todo mark block as Map
	default:
		d.error(d.pos(tok), "unexpected token at top level, got %v", tok)
	}
}

func (d *Decoder) rule(tok *token) {
	d.mss.pushBlock()
	d.selectors(tok)
	d.expect(tokenLBrace)
	d.block()
	d.mss.popBlock()
}

func (d *Decoder) block() {
	for {
		tok := d.next()
		switch tok.t {
		case tokenHash, tokenAttachment, tokenClass, tokenLBracket:
			d.rule(tok)
		case tokenIdent, tokenInstance:
			keyword := tok.value
			if tok.t == tokenInstance {
				d.mss.setInstance(tok.value[:len(tok.value)-1]) // strip /
				tok = d.next()
				if tok.t != tokenIdent {
					d.error(d.pos(tok), "expected IDENT found %v", tok)
				}
				keyword = tok.value
			}
			d.expect(tokenColon)
			d.expressionList()
			d.mss.setProperty(keyword, d.lastValue,
				position{line: tok.line, column: tok.column, filename: d.filename, filenum: d.filesParsed, index: d.propertyIndex},
			)
			d.propertyIndex += 1
			d.expectEndOfStatement()
		case tokenRBrace:
			return
		default:
			d.error(d.pos(tok), "unexpected token %v", tok)
		}
	}
}

// decode multiple selectors, eg:
//   #foo, #bar[zoom=3]
func (d *Decoder) selectors(tok *token) {
	for {
		if tok.t == tokenHash || tok.t == tokenAttachment || tok.t == tokenClass || tok.t == tokenLBracket {
			d.selector(tok)
			tok = d.next()
			if tok.t == tokenComma {
				tok = d.next() // TODO non-selector after comma?
				if tok.t == tokenLBrace {
					// dangling comma
					d.backup()
					break
				}

				continue
			}
			d.backup()
		} else {
			d.error(d.pos(tok), "expected layer, attachment, class or filter, got %v", tok)
		}
		break
	}
}

// decode single selector, eg:
//   #foo::attachment[filter=foo][zoom>=12]
func (d *Decoder) selector(tok *token) {
	d.mss.pushSelector()
	for {
		switch tok.t {
		case tokenHash:
			d.mss.addLayer(tok.value[1:]) // strip #
		case tokenAttachment:
			d.mss.addAttachment(tok.value[2:]) // strip ::
		case tokenClass:
			d.mss.addClass(tok.value[1:]) // strip .
		case tokenLBracket:
			d.filters(tok)
		}

		tok = d.next()
		if tok.t == tokenHash || tok.t == tokenAttachment || tok.t == tokenClass || tok.t == tokenLBracket {
			continue
		} else {
			d.backup()
			break
		}
	}
}

// decode multiple filters. eg:
//   [filter=foo][zoom>=12]
func (d *Decoder) filters(tok *token) {
	for {
		d.filter()
		tok = d.next()
		if tok.t == tokenLBracket {
			continue
		} else {
			d.backup()
			break
		}
	}
}

// decode single filters. eg:
//   [filter=foo]
func (d *Decoder) filter() {
	tok := d.next()
	if tok.t == tokenIdent && tok.value == "zoom" {
		compOp := d.comp()
		tok = d.next()
		if tok.t != tokenNumber {
			d.error(d.pos(tok), "zoom requires num, got %v", tok)
		}
		level, err := strconv.ParseInt(tok.value, 10, 64)
		if err != nil {
			d.error(d.pos(tok), "invalid zoom level %v: %v", tok, err)
		}
		if compOp == REGEX {
			d.error(d.pos(tok), "regular expressions are not allowed for zoom levels")
		}
		d.mss.addZoom(compOp, level)
		d.expect(tokenRBracket)
		return
	}

	var field string
	switch tok.t {
	case tokenString:
		field = tok.value[1 : len(tok.value)-1]
	case tokenIdent:
		field = tok.value
	default:
		d.error(d.pos(tok), "expected zoom or field name in filter, got '%s'", tok.value)
	}

	compOp := d.comp()
	tok = d.next()
	var value interface{}
	switch tok.t {
	case tokenString:
		value = tok.value[1 : len(tok.value)-1]
	case tokenNumber:
		value, _ = strconv.ParseFloat(tok.value, 64)
	case tokenIdent:
		if tok.value == "null" {
			value = nil
		} else {
			d.error(d.pos(tok), "unexpected value in filter '%s'", tok.value)
		}
	default:
		d.error(d.pos(tok), "unexpected value in filter '%s'", tok.value)
	}
	d.expect(tokenRBracket)
	d.mss.addFilter(field, compOp, value)
}

// decode comparision. eg:
//   = or >=
func (d *Decoder) comp() CompOp {
	tok := d.next()
	if tok.t != tokenComp {
		d.error(d.pos(tok), "expected comparsion, got '%s'", tok.value)
	}
	compOp, err := parseCompOp(tok.value)
	if err != nil {
		d.error(d.pos(tok), "invalid comparsion operator '%s': %v", tok.value, err)
	}
	return compOp
}

// expect consumes the next token checks that it is of type t
func (d *Decoder) expect(t tokenType) {
	if tok := d.next(); tok.t != t {
		d.error(d.pos(tok), "expected %v found %v", t, tok)
	}
}

// expectEndOfStatement checks for semicolon or closing block `}`
func (d *Decoder) expectEndOfStatement() {
	tok := d.next()

	if tok.t == tokenRBrace {
		d.backup()
		return
	}
	d.backup()
	d.expect(tokenSemicolon)
}

func (d *Decoder) expressionList() {
	startTok := d.next()
	d.backup()

	d.expression()
	for {
		tok := d.next()
		if tok.t == tokenComma {
			d.expression()
		} else if tok.t == tokenFunction && d.expr.code[len(d.expr.code)-1].T == typeFunctionEnd {
			// non-comma separated list, only between functions, e.g. raster-colorizer-stops: stop(0, #47443e) stop(50, #77654a);
			d.backup()
			d.expression()
		} else {
			d.backup()
			break
		}
	}

	d.expr.pos = position{line: startTok.line, column: startTok.column, filename: d.filename, filenum: d.filesParsed, index: d.propertyIndex}
	d.propertyIndex += 1
	d.lastValue = d.expr
	d.expr = &expression{}
}

func (d *Decoder) expression() {
	d.exprPart()
}

func (d *Decoder) exprPart() {
	d.mulExpr()

	for {
		tok := d.next()
		if tok.t == tokenPlus {
			d.mulExpr()
			d.expr.addOperator(typeAdd)
		} else if tok.t == tokenMinus {
			d.mulExpr()
			d.expr.addOperator(typeSubtract)
		} else {
			d.backup()
			break
		}
	}
}

func (d *Decoder) mulExpr() {
	d.negOrValue()

	for {
		tok := d.next()
		if tok.t == tokenMultiply {
			d.negOrValue()
			d.expr.addOperator(typeMultiply)
		} else if tok.t == tokenDivide {
			d.negOrValue()
			d.expr.addOperator(typeDivide)
		} else {
			d.backup()
			break
		}
	}
}

func (d *Decoder) negOrValue() {
	tok := d.next()
	if tok.t == tokenMinus {
		tok := d.next()
		d.value(tok)
		d.expr.addOperator(typeNegation)
	} else {
		d.value(tok)
	}
}

var urlPath = regexp.MustCompile(`url\(['"]?(.*?)['"]?\)`) // TODO quote handling is borked, eg url('foo") or url('foo) is matched

func (d *Decoder) value(tok *token) {
	switch tok.t {
	case tokenString:
		d.expr.addValue(tok.value[1:len(tok.value)-1], typeString)
	case tokenNumber:
		v, err := strconv.ParseFloat(tok.value, 64)
		if err != nil {
			d.error(d.pos(tok), "invalid float %v: %s", v, err)
		}
		d.expr.addValue(v, typeNum)
	case tokenPercentage:
		v, err := strconv.ParseFloat(tok.value[:len(tok.value)-1], 64)
		if err != nil {
			d.error(d.pos(tok), "invalid float %v: %s", v, err)
		}
		d.expr.addValue(v, typePercent)
	case tokenIdent:
		switch tok.value {
		case "true":
			d.expr.addValue(true, typeBool)
		case "false":
			d.expr.addValue(false, typeBool)
		case "null":
			d.expr.addValue(nil, typeKeyword)
		default:
			c, err := color.Parse(tok.value)
			if err == nil {
				d.expr.addValue(c, typeColor)
			} else {
				// TODO check for valid keywords
				d.expr.addValue(tok.value, typeKeyword)
			}
		}
	case tokenHash:
		c, err := color.Parse(tok.value)
		if err != nil {
			d.error(d.pos(tok), "%v, got %v", err, tok)
		}
		d.expr.addValue(c, typeColor)
	case tokenAtKeyword:
		d.expr.addValue(tok.value[1:], typeVar)
	case tokenURI:
		match := urlPath.FindStringSubmatch(tok.value)
		d.expr.addValue(match[1], typeURL)
	case tokenLBracket:
		// [field]
		tok = d.next()
		if tok.t != tokenIdent {
			d.error(d.pos(tok), "expected identifier in field name, got %v", tok)
		}
		d.expr.addValue("["+tok.value+"]", typeField)
		d.expect(tokenRBracket)
	case tokenFunction:
		d.expr.addValue(tok.value[:len(tok.value)-1], typeFunction) // strip lparen
		d.functionParams()
	case tokenLParen:
		d.exprPart()
		d.expect(tokenRParen)
	default:
		d.error(d.pos(tok), "unexpected value %v", tok)
	}
}

func (d *Decoder) functionParams() {
	for {
		d.exprPart()
		tok := d.next()
		if tok.t == tokenRParen {
			d.expr.addValue(nil, typeFunctionEnd)
			break
		}
		if tok.t == tokenComma {
			continue
		}
		d.error(d.pos(tok), "expected end of function or comma, got %v", tok)
	}
}

type ParseError struct {
	Filename string
	Line     int
	Column   int
	Err      string
}

func (p *ParseError) Error() string {
	file := p.Filename
	if file == "" {
		file = "?"
	}
	return fmt.Sprintf("%s in %s line: %d col: %d", p.Err, file, p.Line, p.Column)
}

func (d *Decoder) pos(tok *token) position {
	return position{
		filename: d.filename,
		line:     tok.line,
		column:   tok.column,
		filenum:  d.filesParsed,
	}
}

func (d *Decoder) error(pos position, format string, args ...interface{}) {
	panic(&ParseError{
		Filename: pos.filename,
		Line:     pos.line,
		Column:   pos.column,
		Err:      fmt.Sprintf(format, args...),
	})
}

func (d *Decoder) warn(pos position, format string, args ...interface{}) {
	d.warnings = append(d.warnings,
		warning{
			file: pos.filename,
			line: pos.line,
			col:  pos.column,
			msg:  fmt.Sprintf(format, args...),
		},
	)
}
