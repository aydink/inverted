package inverted

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeywordAnalyzer(t *testing.T) {
	kt := NewKeywordTokenizer()
	a := NewSimpleAnalyzer(kt)

	want := []Token{{0, 12, 0, "Hello World!"}}
	got := a.Analyze("Hello World!")

	assert.EqualValues(t, want, got)
}

func TestMaxTokenLengthFilter(t *testing.T) {

	simpleTokenizer := NewSimpleTokenizer()
	simpleAnalyzer := NewSimpleAnalyzer(simpleTokenizer)
	maxLengthFilter := NewMaxTokenLengthFilter(5)

	simpleAnalyzer.AddTokenFilter(maxLengthFilter)

	text := "aydın verylongtoken short token"

	want := []Token{
		{0, 6, 0, "aydın"},
		{7, 20, 1, "veryl"},
		{21, 26, 2, "short"},
		{27, 32, 3, "token"},
	}
	got := simpleAnalyzer.Analyze(text)
	t.Log(got)

	assert.EqualValues(t, want, got)
}
