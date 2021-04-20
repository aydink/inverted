package main

import (
	"bufio"
	"compress/bzip2"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type Sentence struct {
	text     string
	category []string
}

func (s *Sentence) Text() string {
	return s.text
}

func (s *Sentence) Category() []string {
	return s.category
}

func IndexSentence() {

	//f, err := os.Open("data/tur_sentences.tsv.bz2")
	f, err := os.Open("data/sentences.tar.bz2")
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer f.Close()

	gr := bzip2.NewReader(f)

	counter := 0
	scanner := bufio.NewScanner(gr)
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "\t")
		if len(line) == 3 {
			sentence := &Sentence{}
			sentence.text = line[2]
			sentence.category = append(sentence.category, line[1])
			idx.Add(sentence)

			if (counter % 10000) == 0 {
				fmt.Println(counter, sentence)
			}
			counter++
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}

var idx *InvertedIndex

func main() {

	analyzer := NewSimpleAnalyzer(NewSimpleTokenizer())
	analyzer.AddTokenFilter(NewTurkishLowercaseFilter())
	analyzer.AddTokenFilter(NewTurkishAccentFilter())
	analyzer.AddTokenFilter(NewTurkishStemFilter())

	idx = NewInvertedIndex(analyzer)

	IndexSentence()

	idx.updateAvgFieldLen()
	idx.BuildCategoryBitmap()

	/*
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			t := scanner.Text()

			hits := idx.Search(t)
			fmt.Println("Number of hits:", len(hits))
			fmt.Println("-----------------------------")

			if len(hits) > 20 {
				hits = hits[0:20]
			}

			for k, v := range hits {
				fmt.Printf("%d\t%f\t%s\n", k, v.boost, idx.store[v.docId])
			}
		}
		//fmt.Println(idx.tokenStats())


	*/

	http.HandleFunc("/search", SearchHandler)
	http.ListenAndServe(":8080", nil)

}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	hits := idx.Search(q)

	fmt.Fprintln(w, "Number of hits:", len(hits), "<br>")
	fmt.Fprintln(w, "<hr>")

	if len(hits) > 20 {
		hits = hits[0:20]
	}

	for k, v := range hits {
		fmt.Fprintf(w, "%d\t%f\t%s<br>", k, v.boost, idx.store[v.docId])
	}
}
