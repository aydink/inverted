package inverted

import (
	"log"
	"sort"

	"github.com/RoaringBitmap/roaring"
)

type Posting struct {
	DocId     uint32
	frequency uint32
	Boost     float32
	positions []uint32
}

type Term struct {
	Value    string  // string representaion of the Term
	Idf      float32 // Inverse Document Frequency of the Term
	Postings []Posting
}

// The main struct that represent an Inveted Index
type InvertedIndex struct {
	docId   uint32
	NumDocs uint32
	index   map[string][]Posting

	// document categories
	docCategory map[string][]uint32

	// roaring bitmaps to store bookCategory bitmaps
	categoryBitmaps map[string]*roaring.Bitmap

	// store field length in number of tokens
	fieldLen []uint32

	// avarage field length
	avgFieldLen float64

	// Analyzer to use for text analysis and tokenization
	analyzer Analyzer

	// Storage backend
	readOnly bool
}

func NewInvertedIndex(analyzer Analyzer) *InvertedIndex {
	idx := &InvertedIndex{}
	idx.docId = 0

	idx.index = make(map[string][]Posting)

	// document categories
	idx.docCategory = make(map[string][]uint32)

	idx.categoryBitmaps = make(map[string]*roaring.Bitmap)

	// store field length in number of tokens
	idx.fieldLen = make([]uint32, 0)

	idx.analyzer = analyzer

	// this is an in memory index
	idx.readOnly = false
	return idx
}

func (idx *InvertedIndex) Add(doc string, categories []string) uint32 {

	if idx.readOnly {
		log.Fatalln("The index is in read only mode!")
	}
	// store docId as return value
	docId := idx.docId

	// Start the analysis process
	tokens := idx.analyzer.Analyze(doc)

	for key, val := range tokenPositions(tokens) {
		//fmt.Println(key, val)
		posting := Posting{idx.docId, uint32(len(val)), 1.0, val}
		idx.index[key] = append(idx.index[key], posting)
	}

	// add document categories to index
	for _, category := range categories {
		idx.docCategory[category] = append(idx.docCategory[category], idx.docId)
	}

	// increment docId after ever document
	idx.docId++

	idx.fieldLen = append(idx.fieldLen, uint32(len(tokens)))

	// increment total number of documents in index
	idx.NumDocs++

	return docId
}

func (idx *InvertedIndex) UpdateAvgFieldLen() {
	total := 0

	for _, v := range idx.fieldLen {
		total += int(v)
	}

	idx.avgFieldLen = float64(total) / float64(idx.NumDocs)
}

func (idx *InvertedIndex) scorePosting(postings []Posting) {
	//fmt.Println(postings)
	for i := range postings {
		postings[i].Boost = float32(idf(float64(len(postings)), float64(idx.NumDocs)) * tf(float64(postings[i].frequency), float64(idx.fieldLen[postings[i].DocId]), idx.avgFieldLen))
		//fmt.Println(postings[i].boost)
	}
	//fmt.Println(postings)
}

func (idx *InvertedIndex) BuildCategoryBitmap() {

	for k, v := range idx.docCategory {
		rb := roaring.NewBitmap()
		rb.AddMany(v)
		idx.categoryBitmaps[k] = rb
	}
}

func (idx *InvertedIndex) GetFacetCounts(postings []Posting) []FacetCount {
	facetCounts := make([]FacetCount, 0)

	rb := roaring.NewBitmap()
	for _, posting := range postings {
		rb.Add(posting.DocId)
	}

	for k, v := range idx.categoryBitmaps {
		fc := FacetCount{}
		fc.Name = k
		fc.Count = int(v.AndCardinality(rb))

		// add only if facet count is not zero
		if fc.Count > 0 {
			facetCounts = append(facetCounts, fc)
		}
	}

	sort.Sort(byFacetCount(facetCounts))

	return facetCounts
}

func (idx *InvertedIndex) FacetFilter(postings []Posting, category string) []Posting {

	result := make([]Posting, 0)
	rb := idx.categoryBitmaps[category]

	for _, posting := range postings {
		if rb.Contains(posting.DocId) {
			result = append(result, posting)
		}
	}
	return result
}

func (idx *InvertedIndex) Filter(category string) *roaring.Bitmap {

	if val, ok := idx.categoryBitmaps[category]; ok {
		return val.Clone()
	}

	return roaring.NewBitmap()
}

func (idx *InvertedIndex) TokenStats() []FacetCount {

	stats := make([]FacetCount, 0)

	for k, v := range idx.index {
		fc := FacetCount{}
		fc.Name = k

		counter := 0
		for _, posting := range v {
			counter += int(posting.frequency)
		}

		fc.Count = counter
		stats = append(stats, fc)
	}

	sort.Sort(byFacetCount(stats))
	return stats
}

// tokenPositions calculate position data for each token
func tokenPositions(tokens []Token) map[string][]uint32 {
	tp := make(map[string][]uint32)

	for i := range tokens {
		tp[tokens[i].value] = append(tp[tokens[i].value], tokens[i].position)
	}

	return tp
}

func (idx *InvertedIndex) CalculateIndexSize() {

	numPosting := 0
	numPositions := 0

	for _, v := range idx.index {
		numPosting += len(v)
		for _, p := range v {
			numPositions += len(p.positions)
		}
	}

	ramPosting := numPosting * 40
	ramPositions := numPositions * 4

	log.Printf("numPosting:%d, numPositions:%d", numPosting, numPositions)
	log.Printf("ramPosting:%d, ramPositions:%d", ramPosting, ramPositions)
}

func (idx *InvertedIndex) AnalyzeText(value string) []string {

	tokens := idx.analyzer.Analyze(value)
	s := make([]string, 0)
	for _, token := range tokens {
		if token.value != "" {
			s = append(s, token.value)
		}
	}

	return s
}
