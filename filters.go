package inverted

import (
	"bufio"
	"compress/gzip"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/reiver/go-porterstemmer"
)

// TurkishLowercaseFilter lowercases all tokens
// respecting Tukish special lowercase rules like "İ"->"i", "I"->"ı"
type turkishLowercaseFilter struct{}

func NewTurkishLowercaseFilter() TokenFilterer {
	filter := turkishLowercaseFilter{}
	return filter
}

func (tf turkishLowercaseFilter) Filter(tokens []Token) []Token {
	for i := range tokens {
		tokens[i].value = strings.ToLowerSpecial(unicode.TurkishCase, tokens[i].value)
	}
	return tokens
}

type turkishAccentFilter struct{}

func NewTurkishAccentFilter() TokenFilterer {
	filter := turkishAccentFilter{}
	return filter
}

/* Filter replaces Turkish accented chracters with not accented versions
"â" -> "a"
"î" -> "i"
"û" -> "u"
"Â" -> "A"
"Î" -> "İ"
"Û" -> "U"
*/
func (tf turkishAccentFilter) Filter(tokens []Token) []Token {
	replacer := strings.NewReplacer("â", "a", "î", "i", "û", "u", "Â", "A", "Î", "İ", "Û", "U")
	for i := range tokens {
		tokens[i].value = replacer.Replace(tokens[i].value)
	}
	return tokens
}

type turkishStemFilter struct {
	dict map[string]string
}

// NewTurkishStemFilter loads 1.087.312 Turkish words and their stems into a map
// it uses a simple map lookup to find stem of a token, if any
func NewTurkishStemFilter() TokenFilterer {
	filter := turkishStemFilter{}
	filter.dict = loadTurkishStems()
	log.Println("Turkish stemmer dictionary loaded:", len(filter.dict), "items")
	return filter
}

func (tf turkishStemFilter) Filter(tokens []Token) []Token {

	for i := range tokens {
		if val, ok := tf.dict[tokens[i].value]; ok {
			tokens[i].value = val
		}
	}
	return tokens
}

func loadTurkishStems() map[string]string {

	f, err := os.Open("data/turkish_synonym.txt.gz")
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer gr.Close()

	dict := make(map[string]string)

	scanner := bufio.NewScanner(gr)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "=>")
		dict[line[0]] = line[1]
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return dict
}

type englishStemFilter struct{}

// NewEnglishStemFilter creates a Porter Stemmer for english tokens
func NewEnglishStemFilter() TokenFilterer {
	filter := englishStemFilter{}
	return filter
}

func (tf englishStemFilter) Filter(tokens []Token) []Token {

	for i := range tokens {
		tokens[i].value = porterstemmer.StemString(tokens[i].value)
	}
	return tokens
}

type stopFilter struct {
	list map[string]bool
}

func NewStopFilter(stopList []string) TokenFilterer {
	filter := &stopFilter{}
	filter.list = make(map[string]bool)

	for _, v := range stopList {
		filter.list[v] = true
	}

	return filter
}

func (tf *stopFilter) Filter(tokens []Token) []Token {
	s := make([]Token, 0)

	for i := range tokens {
		if _, ok := tf.list[tokens[i].value]; !ok {
			s = append(s, tokens[i])
		}
	}
	return s
}

type maxTokenLengthFilter struct {
	maxTokenLength int
}

func NewMaxTokenLengthFilter(maxTokenLength int) TokenFilterer {
	filter := &maxTokenLengthFilter{maxTokenLength}
	return filter
}

func (tf *maxTokenLengthFilter) Filter(tokens []Token) []Token {

	for i := range tokens {
		r := []rune(tokens[i].value)
		if len(r) > tf.maxTokenLength {
			r = r[:tf.maxTokenLength]
			tokens[i].value = string(r)
		}
	}

	return tokens
}
