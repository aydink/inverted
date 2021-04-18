package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

// TurkishLowercaseFilter lowercases all tokens respecting Tukish special lowercase rules like "İ"->"i", "I"->"ı"
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

func (tf turkishAccentFilter) Filter(tokens []Token) []Token {
	replacer := strings.NewReplacer("â", "a", "î", "i", "û", "u", "Â", "A", "Î", "İ", "Û", "U")
	for i := range tokens {
		tokens[i].value = replacer.Replace(tokens[i].value)
	}
	return tokens
}

type turkishStemFilter struct{}

func NewTurkishStemFilter() TokenFilterer {
	filter := turkishStemFilter{}
	return filter
}

func (tf turkishStemFilter) Filter(tokens []Token) []Token {

	for i := range tokens {
		if val, ok := dict[tokens[i].value]; ok {
			tokens[i].value = val
		}
	}
	return tokens
}

var dict map[string]string

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

func init() {
	dict = loadTurkishStems()
	fmt.Println("Turkish stemmer dictionary loaded:", len(dict), "items")
}
