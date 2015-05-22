package mapserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItem(t *testing.T) {
	assert.Equal(t, `KEY "str"`, Item{"key", quote("str")}.String())
	assert.Equal(t, `"str"`, Item{"", quote("str")}.String())
	assert.Equal(t, `42`, Item{"", 42}.String())
	assert.Equal(t, `KEY 42`, Item{"key", 42}.String())

	assert.Equal(t, `FOO ON`, Item{"foo", "ON"}.String())

	// TODO
	// assert.Equal(t, `"quote\""`, Item{"", "quote\""}.String())
}

func TestBlock(t *testing.T) {
	assert.Equal(t, `KEY "str"`, Block{"", []Item{{"key", quote("str")}}}.String())
	assert.Equal(t,
		`KEY "str"
KEY "str"`,
		Block{"", []Item{{"key", quote("str")}, {"key", quote("str")}}}.String())

	assert.Equal(t,
		`CLASS
  KEY "str"
  KEY "str"
END`,
		Block{"CLASS", []Item{{"key", quote("str")}, {"key", quote("str")}}}.String())

	assert.Equal(t,
		`CLASS
  KEY "str"
  LABEL
    FOO 42
  END
END`,
		Block{"CLASS",
			[]Item{
				{"key", quote("str")},
				{"", Block{
					"label",
					[]Item{
						{"foo", 42},
					},
				}},
			},
		}.String())
}
