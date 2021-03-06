package inverted

import (
	"math"
	"os"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func uint32ToBytes(x uint32) []byte {
	var buf [4]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	return buf[:]
}

func bytesToUint32le(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func int32ToBytes(x int32) []byte {
	var buf [4]byte
	buf[0] = byte(x >> 0)
	buf[1] = byte(x >> 8)
	buf[2] = byte(x >> 16)
	buf[3] = byte(x >> 24)
	return buf[:]
}

func bytesToInt32le(b []byte) int32 {
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func float64ToBytes(f float64) []byte {
	var buf [8]byte
	n := math.Float64bits(f)
	buf[0] = byte(n >> 0)
	buf[1] = byte(n >> 8)
	buf[2] = byte(n >> 16)
	buf[3] = byte(n >> 24)
	buf[4] = byte(n >> 32)
	buf[5] = byte(n >> 40)
	buf[6] = byte(n >> 48)
	buf[7] = byte(n >> 56)
	return buf[:]
}

func bytesToFloat64(b []byte) float64 {
	n := uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	return math.Float64frombits(n)
}

func abs(x int) int {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}

func BinarySearch(p []int, value int) int {

	start_index := 0
	end_index := len(p) - 1

	for start_index <= end_index {

		median := (start_index + end_index) / 2

		if p[median] < value {
			start_index = median + 1
		} else {
			end_index = median - 1
		}
	}

	if start_index == len(p) || p[start_index] != value {
		return -1
	} else {
		return start_index
	}
}

func BinarySearchRange(p []int, value int) int {

	start_index := 0
	end_index := len(p) - 1

	for start_index <= end_index {

		median := (start_index + end_index) / 2

		if p[median] < value {
			start_index = median + 1
		} else {
			end_index = median - 1
		}
	}

	if start_index == len(p) {
		return -1
	}

	if p[start_index] != value {
		return end_index
	} else {
		return start_index
	}
}

/**
 * Performs binary search to find the first phrase that matches the prefix
 * @param prefix the desired prefix
 * @return ths index of the first matching phrase
 */
func FindFirst(p []Term, prefix string) int {
	low := 0
	high := len(p) - 1
	for low <= high {
		mid := (low + high) / 2
		if strings.HasPrefix(p[mid].Value, prefix) {
			if mid == 0 || !strings.HasPrefix(p[mid-1].Value, prefix) {
				return mid
			} else {
				high = mid - 1
			}
		} else if p[mid].Value < prefix {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low
}

/**
 * Performs binary search to find the last phrase that matches the prefix
 * @param prefix the desired prefix
 * @return ths index of the last matching phrase
 */
func FindLast(p []Term, prefix string) int {
	low := 0
	high := len(p) - 1
	for low <= high {
		mid := (low + high) / 2
		if strings.HasPrefix(p[mid].Value, prefix) {
			if mid == len(p)-1 || !strings.HasPrefix(p[mid+1].Value, prefix) {
				return mid
			} else {
				low = mid + 1
			}
		} else if p[mid].Value < prefix {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return high
}

func Intersection2(arr1, arr2 []Posting) []Posting {
	m := len(arr1)
	n := len(arr2)

	min := 0
	if m < n {
		min = m
	} else {
		min = n
	}

	p := make([]Posting, 0, min/4)

	i, j := 0, 0

	for i < m && j < n {
		if arr1[i].DocId < arr2[j].DocId {
			i++
		} else if arr2[j].DocId < arr1[i].DocId {
			j++
		} else { /* if arr1[i] == arr2[j] */
			//fmt.Printf(" %d ", arr2[j])
			p = append(p, arr2[j])
			j++
			i++
		}
	}

	return p
}

func Intersection(arr1, arr2 []Posting) []Posting {
	m := len(arr1)
	n := len(arr2)

	min := 0
	if m < n {
		min = m
	} else {
		min = n
	}

	p := make([]Posting, 0, min/4)

	i, j := 0, 0

	for i < m && j < n {
		if arr1[i].DocId < arr2[j].DocId {
			i++
		} else if arr2[j].DocId < arr1[i].DocId {
			j++
		} else {
			arr2[j].Boost += arr1[i].Boost
			p = append(p, arr2[j])
			j++
			i++
		}
	}

	return p
}

func Union(arr1 []Posting, arr2 []Posting) []Posting {
	m := len(arr1)
	n := len(arr2)

	i, j := 0, 0

	p := make([]Posting, 0, m)

	for i < m && j < n {
		if arr1[i].DocId < arr2[j].DocId {
			p = append(p, arr1[i])
			i++
		} else if arr2[j].DocId < arr1[i].DocId {
			p = append(p, arr2[j])
			j++
		} else {
			arr2[j].Boost += arr1[i].Boost
			p = append(p, arr2[j])
			j++
			i++
		}
	}

	/* Print remaining elements of the larger array */
	for i < m {
		p = append(p, arr1[i])
		i++
	}

	for j < n {
		p = append(p, arr2[j])
		j++
	}

	return p
}

func IntersectionPhraseQuery(p1, p2 []Posting, k int) []Posting {
	m := len(p1)
	n := len(p2)

	min := 0
	if m < n {
		min = m
	} else {
		min = n
	}

	p := make([]Posting, 0, min/4)

	i, j := 0, 0

	for i < m && j < n {
		if p1[i].DocId < p2[j].DocId {
			i++
		} else if p2[j].DocId < p1[i].DocId {
			j++
		} else { /* if p1[i] == p2[j] */
			//fmt.Printf(" %d ", p2[j])
			//p2[j].boost += p1[i].boost
			//p = append(p, p2[j])
			//------------------------------------------------------------

			pp1 := p1[i].positions
			pp2 := p2[j].positions

			m1 := len(p1[i].positions)
			n1 := len(p2[j].positions)

			i1, j1 := 0, 0

			for i1 < m1 && j1 < n1 {
				if abs(int(pp1[i1])-int(pp2[j1])) <= k {
					p = append(p, p2[j])
					break
				} else if pp1[i1] < pp2[j1] {
					i1++
				} else {
					j1++
				}
			}

			//------------------------------------------------------------

			j++
			i++
		}
	}

	return p
}

func PhraseQuery_FullMatch(p1, p2 []Posting) []Posting {
	m := len(p1)
	n := len(p2)

	min := 0
	if m < n {
		min = m
	} else {
		min = n
	}

	p := make([]Posting, 0, min/4)

	i, j := 0, 0

	for i < m && j < n {
		if p1[i].DocId < p2[j].DocId {
			i++
		} else if p2[j].DocId < p1[i].DocId {
			j++
		} else {

			pp1 := p1[i].positions
			pp2 := p2[j].positions

			m1 := len(p1[i].positions)
			n1 := len(p2[j].positions)

			i1, j1 := 0, 0

			for i1 < m1 && j1 < n1 {
				//if abs(int(pp1[i1])-int(pp2[j1])) <= k {
				if (int(pp1[i1]) - int(pp2[j1])) == -1 {
					p = append(p, p2[j])
					break
				} else if pp1[i1] < pp2[j1] {
					i1++
				} else {
					j1++
				}
			}

			j++
			i++
		}
	}

	return p
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

type ByBoost []Posting

func (s ByBoost) Len() int           { return len(s) }
func (s ByBoost) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s ByBoost) Less(i, j int) bool { return s[i].Boost > s[j].Boost }

type ByValue []Term

func (a ByValue) Len() int           { return len(a) }
func (a ByValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByValue) Less(i, j int) bool { return a[i].Value < a[j].Value }

func TurkishStringComparer() *collate.Collator {
	col := collate.New(language.Turkish, collate.Numeric, collate.IgnoreCase)
	return col
}
