package inverted

import (
	"errors"
	"log"
	"math"
	"strconv"

	"github.com/RoaringBitmap/roaring"
	"github.com/colinmarc/cdb"
)

func serializePostings(postings []Posting) []byte {
	var sizeInBytes uint32 = 0

	for _, value := range postings {
		// 4 bytes -> DocId
		// 4 bytes -> frequency
		// 4 bytes -> Boost
		// 4 bytes for each term positions
		sizeInBytes += 12 + (value.frequency * 4)
	}

	buf := make([]byte, sizeInBytes)
	//log.Printf("buffer size=%d\n", sizeInBytes)

	cursor := 0

	for _, v := range postings {
		buf[cursor+0] = byte(v.DocId >> 0)
		buf[cursor+1] = byte(v.DocId >> 8)
		buf[cursor+2] = byte(v.DocId >> 16)
		buf[cursor+3] = byte(v.DocId >> 24)
		cursor += 4

		buf[cursor+0] = byte(v.frequency >> 0)
		buf[cursor+1] = byte(v.frequency >> 8)
		buf[cursor+2] = byte(v.frequency >> 16)
		buf[cursor+3] = byte(v.frequency >> 24)
		cursor += 4

		boost := math.Float32bits(v.Boost)
		buf[cursor+0] = byte(boost >> 0)
		buf[cursor+1] = byte(boost >> 8)
		buf[cursor+2] = byte(boost >> 16)
		buf[cursor+3] = byte(boost >> 24)
		cursor += 4
		//fmt.Printf("counter=%d\n", cursor)

		for _, p := range v.positions {
			buf[cursor+0] = byte(p >> 0)
			buf[cursor+1] = byte(p >> 8)
			buf[cursor+2] = byte(p >> 16)
			buf[cursor+3] = byte(p >> 24)
			cursor += 4
		}
	}

	return buf
}

func deserializePostings(buf []byte) ([]Posting, error) {

	//fmt.Printf("size of buffer=%d\n", len(buf))
	postings := make([]Posting, 0)
	cursor := 0

	if len(buf) < 16 {
		return nil, errors.New("byte array is too small: " + strconv.Itoa(len(buf)) + " bytes")
	}

	for {
		if cursor >= len(buf) {
			break
		}

		posting := Posting{}

		// 4 bytes -> DocId
		posting.DocId = uint32(buf[cursor+0]) | uint32(buf[cursor+1])<<8 | uint32(buf[cursor+2])<<16 | uint32(buf[cursor+3])<<24
		cursor += 4

		// 4 bytes -> frequency
		posting.frequency = uint32(buf[cursor+0]) | uint32(buf[cursor+1])<<8 | uint32(buf[cursor+2])<<16 | uint32(buf[cursor+3])<<24
		cursor += 4

		// 4 bytes -> Boost
		posting.Boost = math.Float32frombits(uint32(buf[cursor+0]) | uint32(buf[cursor+1])<<8 | uint32(buf[cursor+2])<<16 | uint32(buf[cursor+3])<<24)
		cursor += 4

		posting.positions = make([]uint32, posting.frequency)

		for i := 0; i < int(posting.frequency); i++ {
			posting.positions[i] = uint32(buf[cursor+0]) | uint32(buf[cursor+1])<<8 | uint32(buf[cursor+2])<<16 | uint32(buf[cursor+3])<<24
			cursor += 4
		}

		postings = append(postings, posting)
	}

	//fmt.Printf("number of postings=%d\n", len(postings))
	return postings, nil
}

// Marshall inverted index to CDB database
func (idx *InvertedIndex) MarshalIndex() error {
	if idx.readOnly {
		log.Println("index is in 'read only' mode hence cannot be marshalled to disk")
		return errors.New("index is in 'read only' mode")
	}

	// update index statitistics and make sure
	// document categories are updated
	idx.UpdateAvgFieldLen()
	idx.BuildCategoryBitmap()

	err := idx.serializeIndex()
	if err != nil {
		log.Println(err)
		return err
	}

	err = idx.serializeDocumentCategories()
	if err != nil {
		log.Println(err)
		return err
	}

	err = idx.serializeIndexMetadata()
	if err != nil {
		log.Println(err)
		return err
	}

	// use committed flag to signal if index committed to disk
	idx.commited = true

	return nil
}

// Serialize term=>postings dictionary to CDB database
func (idx *InvertedIndex) serializeIndex() error {

	writer, err := cdb.Create("data/index.cdb")
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range idx.index {
		buf := serializePostings(v)
		writer.Put([]byte(k), buf)
	}

	writer.Freeze()
	writer.Close()

	return nil
}

// Serialize term=>postings dictionary to CDB database
func (idx *InvertedIndex) serializeIndexMetadata() error {

	writer, err := cdb.Create("data/metadata.cdb")
	if err != nil {
		log.Fatal(err)
	}

	// Now serialize other index properties to CDB file as key => value pair
	// in order to differenciate terms and properties, properties are prepended with a colon ":"

	buf := uint32ToBytes(idx.docId)
	writer.Put([]byte(":docId"), buf)

	buf = uint32ToBytes(idx.NumDocs)
	writer.Put([]byte(":NumDocs"), buf)

	buf = float64ToBytes(idx.avgFieldLen)
	writer.Put([]byte(":avgFieldLen"), buf)
	log.Printf("avgFieldLen=%f\n", idx.avgFieldLen)

	buf = idx.serializeFieldLen()
	writer.Put([]byte(":fieldLen"), buf)

	writer.Freeze()
	writer.Close()

	return nil
}

func ReadPosting_Cdb(term string) []Posting {

	reader, err := cdb.Open("data/index.cdb")
	if err != nil {
		log.Println(err)
	}

	defer reader.Close()

	buf, err := reader.Get([]byte(term))
	if err != nil {
		log.Println(err)
	}

	postings, err := deserializePostings(buf)
	if err != nil {
		log.Println(err)
	}

	return postings
}

func ReadDocument_Cdb(docId uint32) (string, error) {

	reader, err := cdb.Open("data/document.cdb")
	if err != nil {
		log.Println(err)
	}

	defer reader.Close()

	buf, err := reader.Get([]byte(uint32ToBytes(docId)))
	if err != nil {
		log.Println(err)
		return "", err
	}

	return string(buf), nil
}

func loadTermDictionary() (map[string][]Posting, error) {

	index := make(map[string][]Posting)

	reader, err := cdb.Open("data/index.cdb")
	if err != nil {
		log.Println(err)
	}

	defer reader.Close()

	iter := reader.Iter()
	for iter.Next() {
		postings, err := deserializePostings(iter.Value())
		if err != nil {
			log.Println(err)

			return index, err
		}

		index[string(iter.Key())] = postings
	}

	return index, nil
}

func (idx *InvertedIndex) LoadIndexMetadata() error {

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

	return nil
}

// Marshall term=>postings dictionary to CDB database
func (idx *InvertedIndex) serializeDocumentCategories() error {
	writer, err := cdb.Create("data/categories.cdb")

	if err != nil {
		log.Fatal(err)
		return err
	}

	for key, value := range idx.categoryBitmaps {

		buf, err := value.ToBytes()
		if err != nil {
			log.Panicln(err)
			return err
		}
		writer.Put([]byte(key), buf)
	}

	writer.Freeze()
	writer.Close()

	return nil
}

// Marshall term=>postings dictionary to CDB database
func deserializeDocumentCategories() (map[string]*roaring.Bitmap, error) {

	reader, err := cdb.Open("data/categories.cdb")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer reader.Close()

	categoryBitmaps := make(map[string]*roaring.Bitmap)

	iter := reader.Iter()
	for iter.Next() {
		rb := roaring.New()
		_, err = rb.FromBuffer(iter.Value())
		if err != nil {
			log.Println(err)
			return nil, err
		}

		categoryBitmaps[string(iter.Key())] = rb
	}

	return categoryBitmaps, nil
}

func (idx *InvertedIndex) serializeFieldLen() []byte {
	buf := make([]byte, len(idx.fieldLen)*4)

	cursor := 0
	for _, v := range idx.fieldLen {
		buf[cursor+0] = byte(v >> 0)
		buf[cursor+1] = byte(v >> 8)
		buf[cursor+2] = byte(v >> 16)
		buf[cursor+3] = byte(v >> 24)
		cursor += 4
	}

	return buf
}

func deserializeFieldLen(buf []byte) []uint32 {

	fieldLen := make([]uint32, len(buf)/4)

	cursor := 0
	index := 0

	for {
		if cursor >= len(buf) {
			break
		}

		fieldLen[index] = uint32(buf[cursor+0]) | uint32(buf[cursor+1])<<8 | uint32(buf[cursor+2])<<16 | uint32(buf[cursor+3])<<24
		cursor += 4
		index += 1
	}

	return fieldLen
}
