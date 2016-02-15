/*
The CartoCSS scanner is based on the github.com/gorilla/css CSS scanner.

Copyright (c) 2015, Omniscale
Copyright (c) 2013, Gorilla web toolkit
All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

  Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

  Redistributions in binary form must reproduce the above copyright notice, this
  list of conditions and the following disclaimer in the documentation and/or
  other materials provided with the distribution.

  Neither the name of the {organization} nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package mss

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// tokenType identifies the type of lexical tokens.
type tokenType int

// String returns a string representation of the token type.
func (t tokenType) String() string {
	return tokenNames[t]
}

// token represents a token and the corresponding string.
type token struct {
	t      tokenType
	value  string
	line   int
	column int
}

// String returns a string representation of the token.
func (t *token) String() string {
	if len(t.value) > 10 {
		return fmt.Sprintf("%s (line: %d, column: %d): %.10q...",
			t.t, t.line, t.column, t.value)
	}
	return fmt.Sprintf("%s (line: %d, column: %d): %q",
		t.t, t.line, t.column, t.value)
}

// The complete list of tokens in MSS.
const (
	// Scanner flags
	tokenError tokenType = iota
	tokenEOF
	// regular tokens
	tokenIdent
	tokenAtKeyword
	tokenString
	tokenHash
	tokenAttachment
	tokenClass
	tokenInstance
	tokenLBrace
	tokenRBrace
	tokenLBracket
	tokenRBracket
	tokenLParen
	tokenRParen
	tokenColon
	tokenSemicolon
	tokenComma
	tokenPlus
	tokenMinus
	tokenMultiply
	tokenDivide
	tokenComp
	tokenNumber
	tokenPercentage
	tokenDimension
	tokenURI
	tokenUnicodeRange
	tokenS
	tokenComment
	tokenFunction
	tokenIncludes
	tokenDashMatch
	tokenPrefixMatch
	tokenSuffixMatch
	tokenSubstringMatch
	tokenChar
	tokenBOM
)

// tokenNames maps tokenType's to their names. Used for conversion to string.
var tokenNames = map[tokenType]string{
	tokenError:          "error",
	tokenEOF:            "EOF",
	tokenIdent:          "IDENT",
	tokenAtKeyword:      "ATKEYWORD",
	tokenString:         "STRING",
	tokenHash:           "HASH",
	tokenAttachment:     "ATTACHMENT",
	tokenClass:          "CLASS",
	tokenInstance:       "INSTANCE",
	tokenLBrace:         "LBRACE",
	tokenRBrace:         "RBRACE",
	tokenLBracket:       "LBRACKET",
	tokenRBracket:       "RBRACKET",
	tokenLParen:         "LPAREN",
	tokenRParen:         "RPAREN",
	tokenColon:          "COLON",
	tokenSemicolon:      "SEMICOLON",
	tokenComma:          "COMMA",
	tokenPlus:           "PLUS",
	tokenMinus:          "MINUS",
	tokenMultiply:       "MULTIPLY",
	tokenDivide:         "DIVIDE",
	tokenComp:           "COMP",
	tokenNumber:         "NUMBER",
	tokenPercentage:     "PERCENTAGE",
	tokenDimension:      "DIMENSION",
	tokenURI:            "URI",
	tokenUnicodeRange:   "UNICODE-RANGE",
	tokenS:              "S",
	tokenComment:        "COMMENT",
	tokenFunction:       "FUNCTION",
	tokenIncludes:       "INCLUDES",
	tokenDashMatch:      "DASHMATCH",
	tokenPrefixMatch:    "PREFIXMATCH",
	tokenSuffixMatch:    "SUFFIXMATCH",
	tokenSubstringMatch: "SUBSTRINGMATCH",
	tokenChar:           "CHAR",
	tokenBOM:            "BOM",
}

// Macros and productions -----------------------------------------------------
// http://www.w3.org/TR/css3-syntax/#tokenization

var macroRegexp = regexp.MustCompile(`\{[a-z]+\}`)

// macros maps macro names to patterns to be expanded.
var macros = map[string]string{
	// must be escaped: `\.+*?()|[]{}^$`
	"ident":      `-?{nmstart}{nmchar}*`,
	"name":       `{nmchar}+`,
	"nmstart":    `[a-zA-Z_]|{nonascii}|{escape}`,
	"nonascii":   "[\u0080-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]",
	"unicode":    `\\[0-9a-fA-F]{1,6}{wc}?`,
	"escape":     "{unicode}|\\\\[\u0020-\u007E\u0080-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]",
	"nmchar":     `[a-zA-Z0-9_-]|{nonascii}|{escape}`,
	"num":        `-?[0-9]*\.?[0-9]+`,
	"string":     `"(?:{stringchar}|')*"|'(?:{stringchar}|")*'`,
	"stringchar": `{urlchar}|[ ]|\\{nl}`,
	"urlchar":    "[\u0009\u0021\u0023-\u0026\u0028-\u007E]|{nonascii}|{escape}",
	"nl":         `[\n\r\f]|\r\n`,
	"w":          `{wc}*`,
	"wc":         `[\t\n\f\r ]`,
}

// productions maps the list of tokens to patterns to be expanded.
var productions = map[tokenType]string{
	// Unused regexps (matched using other methods) are commented out.
	tokenIdent:        `{ident}`,
	tokenAtKeyword:    `@{ident}`,
	tokenString:       `{string}`,
	tokenHash:         `#{name}`,
	tokenAttachment:   `::{name}`,
	tokenClass:        `\.{name}`,
	tokenInstance:     `{ident}/`,
	tokenNumber:       `{num}`,
	tokenPercentage:   `{num}%`,
	tokenDimension:    `{num}{ident}`,
	tokenURI:          `url\({w}(?:{string}|{urlchar}*){w}\)`,
	tokenUnicodeRange: `U\+[0-9A-F\?]{1,6}(?:-[0-9A-F]{1,6})?`,
	tokenS:            `{wc}+`,
	tokenComment:      `/\*[^\*]*[\*]+(?:[^/][^\*]*[\*]+)*/`,
	tokenFunction:     `{ident}\(`,
	tokenComp:         `>=|<=|>|<|!=|=~|=`,
	//tokenIncludes:       `~=`,
	//tokenDashMatch:      `\|=`,
	//tokenPrefixMatch:    `\^=`,
	//tokenSuffixMatch:    `\$=`,
	//tokenSubstringMatch: `\*=`,
	//tokenChar:           `[^"']`,
	//tokenBOM:            "\uFEFF",
}

// matchers maps the list of tokens to compiled regular expressions.
//
// The map is filled on init() using the macros and productions defined in
// the CSS specification.
var matchers = map[tokenType]*regexp.Regexp{}

// matchOrder is the order to test regexps when first-char shortcuts
// can't be used.
var matchOrder = []tokenType{
	tokenURI,
	tokenFunction,
	tokenUnicodeRange,
	tokenInstance,
	tokenIdent,
	tokenDimension,
	tokenPercentage,
	tokenNumber,
	tokenComp,
}

func init() {
	// replace macros and compile regexps for productions.
	replaceMacro := func(s string) string {
		return "(?:" + macros[s[1:len(s)-1]] + ")"
	}
	for t, s := range productions {
		for macroRegexp.MatchString(s) {
			s = macroRegexp.ReplaceAllStringFunc(s, replaceMacro)
		}
		matchers[t] = regexp.MustCompile("^(?:" + s + ")")
	}
}

// Scanner --------------------------------------------------------------------

// New returns a new CSS scanner for the given input.
func newScanner(input string) *scanner {
	// Normalize newlines.
	input = strings.Replace(input, "\r\n", "\n", -1)
	return &scanner{
		input: input,
		row:   1,
		col:   1,
	}
}

// Scanner scans an input and emits tokens following the CSS3 specification.
type scanner struct {
	input string
	pos   int
	row   int
	col   int
	err   *token
}

// Next returns the next token from the input.
//
// At the end of the input the token type is tokenEOF.
//
// If the input can't be tokenized the token type is tokenError. This occurs
// in case of unclosed quotation marks or comments.
func (s *scanner) Next() *token {
	if s.err != nil {
		return s.err
	}
	if s.pos >= len(s.input) {
		s.err = &token{tokenEOF, "", s.row, s.col}
		return s.err
	}
	if s.pos == 0 {
		// Test BOM only once, at the beginning of the file.
		if strings.HasPrefix(s.input, "\uFEFF") {
			return s.emitSimple(tokenBOM, "\uFEFF")
		}
	}
	// There's a lot we can guess based on the first byte so we'll take a
	// shortcut before testing multiple regexps.
	input := s.input[s.pos:]
	switch input[0] {
	case '\t', '\n', '\f', '\r', ' ':
		// Whitespace.
		return s.emitToken(tokenS, matchers[tokenS].FindString(input))
	case '.':
		// Dot is too common to not have a quick check.
		// We'll test if this is a Char; if it is followed by a number it is a
		// dimension/percentage/number, and this will be matched later.
		if len(input) > 1 && !unicode.IsDigit(rune(input[1])) {
			if match := matchers[tokenClass].FindString(input); match != "" {
				return s.emitSimple(tokenClass, match)
			}
			return s.emitSimple(tokenChar, ".")
		}
	case '#':
		// Another common one: Hash or Char.
		if match := matchers[tokenHash].FindString(input); match != "" {
			return s.emitSimple(tokenHash, match)
		}
		return s.emitSimple(tokenChar, "#")
	case '@':
		// Another common one: AtKeyword or Char.
		if match := matchers[tokenAtKeyword].FindString(input); match != "" {
			return s.emitSimple(tokenAtKeyword, match)
		}
		return s.emitSimple(tokenChar, "@")
	case ':':
		// Another common one: Attachment or Char.
		if match := matchers[tokenAttachment].FindString(input); match != "" {
			return s.emitSimple(tokenAttachment, match)
		}
		return s.emitSimple(tokenColon, ":")
	case '%', '&':
		// More common chars.
		return s.emitSimple(tokenChar, string(input[0]))
	case ',':
		return s.emitSimple(tokenComma, string(input[0]))
	case ';':
		return s.emitSimple(tokenSemicolon, string(input[0]))
	case '(':
		return s.emitSimple(tokenLParen, string(input[0]))
	case ')':
		return s.emitSimple(tokenRParen, string(input[0]))
	case '[':
		return s.emitSimple(tokenLBracket, string(input[0]))
	case ']':
		return s.emitSimple(tokenRBracket, string(input[0]))
	case '{':
		return s.emitSimple(tokenLBrace, string(input[0]))
	case '}':
		return s.emitSimple(tokenRBrace, string(input[0]))
	case '+':
		return s.emitSimple(tokenPlus, string(input[0]))
	case '-':
		if match := matchers[tokenNumber].FindString(input); match != "" {
			return s.emitSimple(tokenNumber, match)
		}
		if match := matchers[tokenFunction].FindString(input); match != "" {
			return s.emitSimple(tokenFunction, match)
		}

		return s.emitSimple(tokenMinus, string(input[0]))
	case '*':
		return s.emitSimple(tokenMultiply, string(input[0]))
	// case '/': handled below
	case '"', '\'':
		// String or error.
		match := matchers[tokenString].FindString(input)
		if match != "" {
			return s.emitToken(tokenString, match)
		} else {
			s.err = &token{tokenError, "unclosed quotation mark", s.row, s.col}
			return s.err
		}
	case '/':
		// Comment, error or Char.
		if len(input) > 1 && input[1] == '*' {
			match := matchers[tokenComment].FindString(input)
			if match != "" {
				return s.emitToken(tokenComment, match)
			} else {
				s.err = &token{tokenError, "unclosed comment", s.row, s.col}
				return s.err
			}
		} else if len(input) > 1 && input[1] == '/' {
			idx := strings.Index(input, "\n")
			if idx < 0 {
				// comment at end of document wihout new line
				idx = len(input)
			}
			return s.emitToken(tokenComment, input[:idx])
		}
		return s.emitSimple(tokenDivide, "/")
	}
	// Test all regexps, in order.
	for _, token := range matchOrder {
		if match := matchers[token].FindString(input); match != "" {
			return s.emitToken(token, match)
		}
	}
	// We already handled unclosed quotation marks and comments,
	// so this can only be a Char.
	r, width := utf8.DecodeRuneInString(input)
	token := &token{tokenChar, string(r), s.row, s.col}
	s.col += width
	s.pos += width
	return token
}

// updatePosition updates input coordinates based on the consumed text.
func (s *scanner) updatePosition(text string) {
	width := utf8.RuneCountInString(text)
	lines := strings.Count(text, "\n")
	s.row += lines
	if lines == 0 {
		s.col += width
	} else {
		s.col = utf8.RuneCountInString(text[strings.LastIndex(text, "\n"):])
	}
	s.pos += len(text)
}

// emitToken returns a token for the string v and updates the scanner position.
func (s *scanner) emitToken(t tokenType, v string) *token {
	token := &token{t, v, s.row, s.col}
	s.updatePosition(v)
	return token
}

// emitSimple returns a token for the string v and updates the scanner
// position in a simplified manner.
//
// The string is known to have only ASCII characters and to not have a newline.
func (s *scanner) emitSimple(t tokenType, v string) *token {
	token := &token{t, v, s.row, s.col}
	s.col += len(v)
	s.pos += len(v)
	return token
}
