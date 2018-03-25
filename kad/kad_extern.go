package kad

import (
	"fmt"
	"hahajing/com"
	"log"
)

// SearchReq x
type SearchReq struct {
	ResCh           chan *SearchRes
	MyKeywordStruct *com.MyKeywordStruct
}

// SearchRes x
type SearchRes struct {
	FileLinks []*com.Ed2kFileLink
}

// Ed2kFileStruct x
type Ed2kFileStruct struct {
	Hash [16]byte

	Name        string
	Size        uint64
	Type        string
	Avail       uint32
	MediaLength uint32

	// Do we need publish info and AICH?
}

// GetEd2kLink x
func (f *Ed2kFileStruct) GetEd2kLink() string {
	return com.GetEd2kLink(f.Name, f.Size, f.Hash[:])
}

// GetPrintStr x
func (f *Ed2kFileStruct) GetPrintStr() string {
	log.Printf("Name: %s, Size: %d, Type: %s, Avail:%d\nEd2k: %s\n", f.Name, f.Size, f.Type, f.Avail, f.GetEd2kLink())

	return fmt.Sprintf("Name: %s, Size: %d, Type: %s, Avail:%d\nEd2k: %s\n", f.Name, f.Size, f.Type, f.Avail, f.GetEd2kLink())
}


