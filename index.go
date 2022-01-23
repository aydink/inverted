package inverted

import (
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

	// docCategories will store parentIds for every document that blongs to a category
	parentIds []uint32

	// document categories
	docCategory map[string][]uint32

	// roaring bitmaps to store bookCategory bitmaps
	categoryBitmaps map[string]*roaring.Bitmap

	// store field length in number of tokens
	fieldLen []int

	// avarage field length
	avgFieldLen float64

	// Analyzer to use for text analysis and tokenization
	analyzer Analyzer
}

func NewInvertedIndex(analyzer Analyzer) *InvertedIndex {
	idx := &InvertedIndex{}
	idx.docId = 0

	idx.index = make(map[string][]Posting)
	idx.parentIds = make([]uint32, 0)

	//idx.postings = make([]uint16, 0)

	// document categories
	idx.docCategory = make(map[string][]uint32)

	idx.categoryBitmaps = make(map[string]*roaring.Bitmap)

	// store field length in number of tokens
	idx.fieldLen = make([]int, 0)

	// store page content for future use
	//idx.store = make([]string, 0)

	idx.analyzer = analyzer
	return idx
}

func (idx *InvertedIndex) Add(doc Document) uint32 {
	// store docId as return value
	docId := idx.docId

	// Start the analysis process
	tokens := idx.analyzer.Analyze(doc.Text())

	for key, val := range tokenPositions(tokens) {
		//fmt.Println(key, val)
		posting := Posting{idx.docId, uint32(len(val)), 1.0, val}
		idx.index[key] = append(idx.index[key], posting)

		//idx.postings = append(idx.postings, val...)

		//increment postingIndex
		//idx.postingIndex += uint32(len(val))
	}

	// add document categories to index
	for _, category := range doc.Category() {
		idx.docCategory[category] = append(idx.docCategory[category], idx.docId)
	}

	// increment docId after ever document
	idx.docId++

	//idx.store = append(idx.store, doc.Text())

	idx.fieldLen = append(idx.fieldLen, len(tokens))

	// increment total number of documents in index
	idx.NumDocs++

	return docId
}

// TODO
func (idx *InvertedIndex) Search(q string) []Posting {
	tokens := idx.analyzer.Analyze(q)

	var result []Posting
	var temp []Posting
	var resultPhrase []Posting

	for i, token := range tokens {
		if i == 0 {
			result = make([]Posting, len(idx.index[token.value]))
			copy(result, idx.index[token.value])
			//fmt.Println(result)
			idx.scorePosting(result)
			//fmt.Println(result)
		} else {
			//temp := idx.index[token.value]
			temp = make([]Posting, len(idx.index[token.value]))
			copy(temp, idx.index[token.value])
			idx.scorePosting(temp)

			// boolean AND query
			result = Intersection(temp, result)
			// boolean OR query
			//result = Union(temp, result)
			// Phrase Query
			//result = PhraseQuery_FullMatch(result, temp)
		}
	}

	for i, token := range tokens {
		if i == 0 {
			resultPhrase = make([]Posting, len(idx.index[token.value]))
			copy(resultPhrase, idx.index[token.value])
			//fmt.Println(result)
			idx.scorePosting(result)
			//fmt.Println(result)
		} else {
			//temp := idx.index[token.value]
			temp = make([]Posting, len(idx.index[token.value]))
			copy(temp, idx.index[token.value])
			idx.scorePosting(temp)

			// boolean AND query
			// result = Intersection(temp, result)
			// boolean OR query
			//result = Union(temp, result)
			// Phrase Query
			resultPhrase = PhraseQuery_FullMatch(resultPhrase, temp)
		}
	}

	result = Union(result, resultPhrase)

	//fmt.Println(result)
	sort.Sort(ByBoost(result))
	//fmt.Println("-------------------------------------------------")
	//fmt.Println(result)

	return result
}

func (idx *InvertedIndex) UpdateAvgFieldLen() {
	total := 0

	for _, v := range idx.fieldLen {
		total += v
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
	//fmt.Printf("%+v\n", facetCounts)

	return facetCounts
}

func (idx *InvertedIndex) FacetFilterCategory(postings []Posting, category string) []Posting {

	result := make([]Posting, 0)
	rb := idx.categoryBitmaps[category]

	for _, posting := range postings {
		if rb.Contains(posting.DocId) {
			result = append(result, posting)
		}
	}
	return result
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
