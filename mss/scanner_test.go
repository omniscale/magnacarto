package mss

import "testing"

type tokVal struct {
	t     tokenType
	value string
}

func assertTokens(t *testing.T, scanner *scanner, tokens []tokVal) {
	for _, expectedTok := range tokens {
		tok := scanner.Next()
		if tok.t != expectedTok.t || tok.value != expectedTok.value {
			t.Fatalf("expected '%s' of type %s, got %v", expectedTok.value, expectedTok.t, tok)
		}
	}
	if tok := scanner.Next(); tok.t != tokenEOF {
		t.Fatalf("expected EOF, got %v", tok)
	}
}

func TestScanner(t *testing.T) {
	var scanner *scanner
	scanner = newScanner(`@foo: 1 + 2`)
	assertTokens(t, scanner, []tokVal{
		{tokenAtKeyword, "@foo"},
		{tokenColon, ":"},
		{tokenS, " "},
		{tokenNumber, "1"},
		{tokenS, " "},
		{tokenPlus, "+"},
		{tokenS, " "},
		{tokenNumber, "2"},
	})

	scanner = newScanner(`#bar::foo[type='foo'] {}`)
	assertTokens(t, scanner, []tokVal{
		{tokenHash, "#bar"},
		{tokenAttachment, "::foo"},
		{tokenLBracket, "["},
		{tokenIdent, "type"},
		{tokenComp, "="},
		{tokenString, "'foo'"},
		{tokenRBracket, "]"},
		{tokenS, " "},
		{tokenLBrace, "{"},
		{tokenRBrace, "}"},
	})

	scanner = newScanner(`//comment`)
	assertTokens(t, scanner, []tokVal{
		{tokenComment, "//comment"},
	})

	scanner = newScanner(`/* comment */`)
	assertTokens(t, scanner, []tokVal{
		{tokenComment, "/* comment */"},
	})

}
