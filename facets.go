package main

type FacetCount struct {
	Name  string
	Count int
}

type byFacetCount []FacetCount

func (f byFacetCount) Len() int {
	return len(f)
}

func (f byFacetCount) Less(i, j int) bool {
	return f[i].Count > f[j].Count
}

func (f byFacetCount) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
