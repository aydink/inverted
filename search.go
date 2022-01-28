package inverted

import "sort"

func (idx *InvertedIndex) Search_Cdb(q string) []Posting {
	tokens := idx.analyzer.Analyze(q)

	var result []Posting
	var resultPhrase []Posting

	postings := make(map[int][]Posting)

	for i, token := range tokens {
		postings[i] = ReadPosting_Cdb(token.value)
		idx.scorePosting(postings[i])
	}

	// Apply AND operation
	for i := range tokens {
		if i == 0 {
			result = postings[0]
		} else {
			result = Intersection(result, postings[i])
		}
	}

	// Apply AND operation
	for i := range tokens {
		if i == 0 {
			resultPhrase = postings[0]
		} else {
			resultPhrase = PhraseQuery_FullMatch(resultPhrase, postings[i])
		}
	}

	idx.scorePosting(resultPhrase)

	result = Union(result, resultPhrase)

	//fmt.Println(result)
	sort.Sort(ByBoost(result))
	//fmt.Println("-------------------------------------------------")
	//fmt.Println(result)

	return result
}

// Default search
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

func (idx *InvertedIndex) Search_Mixed(q string) []Posting {
	tokens := idx.analyzer.Analyze(q)

	var result []Posting
	var resultPhrase []Posting
	var temp []Posting

	postings := make(map[int][]Posting)

	for i, token := range tokens {
		if idx.readOnly {
			postings[i] = ReadPosting_Cdb(token.value)
			idx.scorePosting(postings[i])
		} else {
			postings[i] = make([]Posting, len(idx.index[token.value]))
			copy(postings[i], idx.index[token.value])
			idx.scorePosting(postings[i])
		}
	}

	// Apply AND operation
	for i := range tokens {
		if i == 0 {
			result = postings[i]
		} else {
			result = Intersection(result, postings[i])
		}
	}

	// Apply AND operation
	for i := range tokens {
		if i == 0 {
			resultPhrase = make([]Posting, len(postings[i]))
			copy(resultPhrase, postings[i])
			resultPhrase = resetScore(resultPhrase)
		} else {
			temp = make([]Posting, len(postings[i]))
			copy(temp, postings[i])
			temp = resetScore(temp)
			resultPhrase = PhraseQuery_FullMatch(resultPhrase, temp)
		}
	}

	idx.scorePosting(resultPhrase)

	result = Union(result, resultPhrase)

	//fmt.Println(result)
	sort.Sort(ByBoost(result))
	//fmt.Println("-------------------------------------------------")
	//fmt.Println(result)

	return result
}

func resetScore(postings []Posting) []Posting {
	for i := range postings {
		postings[i].Boost = 1.0
	}
	return postings
}

func (idx *InvertedIndex) Search_Mixed_v2(q string) []Posting {
	tokens := idx.analyzer.Analyze(q)

	var result []Posting
	var resultPhrase []Posting
	var temp []Posting

	postings := make(map[int][]Posting)

	for i, token := range tokens {
		if idx.readOnly {
			postings[i] = ReadPosting_Cdb(token.value)
			idx.scorePosting(postings[i])
		} else {
			postings[i] = make([]Posting, len(idx.index[token.value]))
			copy(postings[i], idx.index[token.value])
			idx.scorePosting(postings[i])
		}
	}

	// Apply AND operation
	for i := range tokens {
		if i == 0 {
			result = postings[i]
		} else {
			result = Intersection(result, postings[i])
		}
	}

	// Apply Phrase query scoring only if more than 1 query term exist
	if len(tokens) > 1 {
		for i := range tokens {
			if i == 0 {
				resultPhrase = make([]Posting, len(postings[i]))
				copy(resultPhrase, postings[i])
				//resultPhrase = resetScore(resultPhrase)
			} else {
				temp = make([]Posting, len(postings[i]))
				copy(temp, postings[i])
				//temp = resetScore(temp)
				resultPhrase = PhraseQuery_FullMatch(resultPhrase, temp)
				result = Union(result, resultPhrase)
			}
		}
		//idx.scorePosting(resultPhrase)
		//result = Union(result, resultPhrase)
	}

	//fmt.Println(result)
	sort.Sort(ByBoost(result))
	//fmt.Println("-------------------------------------------------")
	//fmt.Println(result)

	return result
}
