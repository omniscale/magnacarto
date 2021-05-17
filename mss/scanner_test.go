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
	for _, tt := range []struct {
		text   string
		tokens []tokVal
	}{
		{
			text: `@foo: 1 + 2`,
			tokens: []tokVal{
				{tokenAtKeyword, "@foo"},
				{tokenColon, ":"},
				{tokenS, " "},
				{tokenNumber, "1"},
				{tokenS, " "},
				{tokenPlus, "+"},
				{tokenS, " "},
				{tokenNumber, "2"},
			},
		},
		{
			text: `#bar::foo[type= 'foo'] {}`,
			tokens: []tokVal{
				{tokenHash, "#bar"},
				{tokenAttachment, "::foo"},
				{tokenLBracket, "["},
				{tokenIdent, "type"},
				{tokenComp, "="},
				{tokenS, " "},
				{tokenString, "'foo'"},
				{tokenRBracket, "]"},
				{tokenS, " "},
				{tokenLBrace, "{"},
				{tokenRBrace, "}"},
			},
		},
		{
			text: `lighten(a, 30%)`,
			tokens: []tokVal{
				{tokenFunction, "lighten("},
				{tokenIdent, "a"},
				{tokenComma, ","},
				{tokenS, " "},
				{tokenPercentage, "30%"},
				{tokenRParen, ")"},
			},
		},
		{
			text: `[type % 2 =1] {}`,
			tokens: []tokVal{
				{tokenLBracket, "["},
				{tokenIdent, "type"},
				{tokenS, " "},
				{tokenModulo, "%"},
				{tokenS, " "},
				{tokenNumber, "2"},
				{tokenS, " "},
				{tokenComp, "="},
				{tokenNumber, "1"},
				{tokenRBracket, "]"},
				{tokenS, " "},
				{tokenLBrace, "{"},
				{tokenRBrace, "}"},
			},
		},
		{text: `//comment`, tokens: []tokVal{{tokenComment, "//comment"}}},
		{text: `// comment`, tokens: []tokVal{{tokenComment, "// comment"}}},
		{text: "/* comment\n comment */", tokens: []tokVal{{tokenComment, "/* comment\n comment */"}}},
	} {
		tt := tt
		t.Run("", func(t *testing.T) {
			scanner := newScanner(tt.text)
			assertTokens(t, scanner, tt.tokens)
		})

	}
}
