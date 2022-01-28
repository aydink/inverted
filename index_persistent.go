package inverted

import (
	"log"

	"github.com/colinmarc/cdb"
)

func NewInvertedIndexFromFile(analyzer Analyzer, loadIntoMemory bool) *InvertedIndex {
	idx := &InvertedIndex{}
	idx.docId = 0

	reader, err := cdb.Open("data/metadata.cdb")
	if err != nil {
		log.Println(err)
	}
	defer reader.Close()

	buf, err := reader.Get([]byte(":docId"))
	if err != nil {
		log.Println(err)
	}
	idx.docId = bytesToUint32le(buf)
	log.Printf("docId=%d\n", idx.docId)

	buf, err = reader.Get([]byte(":NumDocs"))
	if err != nil {
		log.Println(err)
	}
	idx.NumDocs = bytesToUint32le(buf)
	log.Printf("NumDocs=%d\n", idx.NumDocs)

	buf, err = reader.Get([]byte(":avgFieldLen"))
	if err != nil {
		log.Println(err)
	}
	idx.avgFieldLen = bytesToFloat64(buf)
	log.Printf("avgFieldLen=%f\n", idx.avgFieldLen)

	buf, err = reader.Get([]byte(":fieldLen"))
	if err != nil {
		log.Println(err)
	}
	idx.fieldLen = deserializeFieldLen(buf)

	if loadIntoMemory {
		termDictionary, err := LoadTermDictionary()
		if err != nil {
			log.Fatalln(err)
		}
		idx.index = termDictionary
	}

	idx.categoryBitmaps, err = deserializeDocumentCategories()
	if err != nil {
		log.Panicln(err)
	}

	// set analyzer
	idx.analyzer = analyzer

	if loadIntoMemory {
		idx.readOnly = false
	} else {
		idx.readOnly = true
	}

	return idx
}
