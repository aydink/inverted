package inverted

import (
	"strings"
)

type SimpleHighlighter struct {
	analyzer Analyzer
}

func NewSimpleHighlighter(analyzer Analyzer) SimpleHighlighter {
	return SimpleHighlighter{analyzer}
}

/*
Do actual text highlighting like <b>term</b>
both document text and query terms are tokenized using the provided Analyzer
*/
func (hl SimpleHighlighter) Highlight(pre, post, text, query string) string {

	queryTokens := idx.analyzer.Analyze(query)
	textTokens := idx.analyzer.Analyze(text)

	//fmt.Println(textTokens)
	//fmt.Println(queryTokens)

	var sb strings.Builder
	cursor := 0
	matched := false

	for _, tt := range textTokens {

		sb.WriteString(text[cursor:tt.start])

		for _, token := range queryTokens {
			if token.value == tt.value {
				sb.WriteString(pre)
				matched = true
			}
		}

		sb.WriteString(text[tt.start:tt.end])

		if matched {
			sb.WriteString(post)
			matched = false
		}
		cursor = int(tt.end)
	}

	//sb.WriteString(text[cursor:])

	return sb.String()
}
