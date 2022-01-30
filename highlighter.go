package inverted

import (
	"strings"
)

type SimpleHighlighter struct {
	analyzer Analyzer
	pre      string
	post     string
}

func NewSimpleHighlighter(analyzer Analyzer) SimpleHighlighter {
	return SimpleHighlighter{analyzer, "<b>", "</b>"}
}

func (hl *SimpleHighlighter) HighlightParams(pre, post string) {
	hl.pre = pre
	hl.post = post
}

/*
Do actual text highlighting like <b>term</b>
both document text and query terms are tokenized using the provided Analyzer
*/
func (hl *SimpleHighlighter) Highlight(document, query string) string {

	queryTokens := hl.analyzer.Analyze(query)
	textTokens := hl.analyzer.Analyze(document)

	//fmt.Println(textTokens)
	//fmt.Println(queryTokens)

	var sb strings.Builder
	cursor := 0
	matched := false

	for _, tt := range textTokens {

		sb.WriteString(document[cursor:tt.start])

		for _, token := range queryTokens {
			if token.value == tt.value {
				sb.WriteString(hl.pre)
				matched = true
			}
		}

		sb.WriteString(document[tt.start:tt.end])

		if matched {
			sb.WriteString(hl.post)
			matched = false
		}
		cursor = int(tt.end)
	}

	//sb.WriteString(document[cursor:])

	return sb.String()
}

type SpanHighlighter struct {
	analyzer      Analyzer
	pre           string
	post          string
	spanDelimeter string
	snippetSize   int
}

func NewSpanHighlighter(analyzer Analyzer) SpanHighlighter {
	return SpanHighlighter{analyzer, "<b>", "</b>", "<br>", 200}
}

func (hl *SpanHighlighter) HighlightParams(pre, post, spanDelimeter string, snippetSize int) {
	hl.pre = pre
	hl.post = post
	hl.spanDelimeter = spanDelimeter
	hl.snippetSize = snippetSize
}

func (hl *SpanHighlighter) Highlight(document string, snippetSize int, query string) string {
	index := NewInvertedIndex(hl.analyzer)

	tokens := index.analyzer.Analyze(document)

	snippets := make([]string, 0)

	start := 0
	for _, token := range tokens {
		if (int(token.end) - start) > snippetSize {
			snippets = append(snippets, document[start:token.end])
			start = int(token.end)
		}
	}

	snippets = append(snippets, document[start:])

	snippetMap := make(map[uint32]string)
	for _, snippet := range snippets {
		docId := index.Add(snippet, nil)
		snippetMap[docId] = snippet
	}

	index.UpdateAvgFieldLen()

	hits := index.SearchOr(query)

	if len(hits) > 2 {
		hits = hits[0:2]
	}

	simpleHighlighter := NewSimpleHighlighter(hl.analyzer)
	hlText := ""

	for _, hit := range hits {
		hlText += simpleHighlighter.Highlight(snippetMap[hit.DocId], query) + hl.spanDelimeter
	}

	return hlText
}
