package main

import (
	"fmt"
)

func main() {

	analyzer := NewSimpleAnalyzer(NewSimpleTokenizer())
	analyzer.AddTokenFilter(NewTurkishLowercaseFilter())
	analyzer.AddTokenFilter(NewTurkishAccentFilter())
	analyzer.AddTokenFilter(NewTurkishStemFilter())

	idx := NewInvertedIndex(analyzer)

	//idx.Add()

	fmt.Scanln()
}
