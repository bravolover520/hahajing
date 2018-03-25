package kad

import (
	"hahajing/com"
	"time"
)

const searchExpires = 5 // 5 seconds for each search living
const ed2kFileReturnNbr = 100
const searchTolerance uint32 = 16777216

// SearchResChSize is size of search result reponse channel for each web search.
const SearchResChSize = 100

// Search x
type Search struct {
	no uint64 // search No.

	resCh chan *SearchRes

	myKeywordStruct *com.MyKeywordStruct // from user and internet

	targetID      ID
	targetKeyword string

	tExpires    int64
	files       []*Ed2kFileStruct
	fileHashMap map[[16]byte]bool

	contacts     []*Contact // contacts in searching target path
	contactIPMap map[uint32]bool
}

func (s *Search) goSearch(contacts []*Contact, pPacketProcessor *PacketProcessor) {
	t := time.Now().Unix()

	for _, pContact := range contacts {
		if s.contactIPMap[pContact.ip] {
			continue
		}

		// send KAD request according to tolerance
		tolerance := s.calcSearchTolerance(pContact)
		if tolerance > searchTolerance {
			if pPacketProcessor.pPacketReqGuard.canPass(t, pContact.ip, kademlia2Req) {
				pPacketProcessor.sendFindValue(pContact, &s.targetID)
			}
		} else {
			if pPacketProcessor.pPacketReqGuard.canPass(t, pContact.ip, kademlia2SearchKeyReq) {
				pPacketProcessor.sendSearchKeyword(pContact, s.targetID.getHash())
			}
		}

		s.contacts = append(s.contacts, pContact)
		s.contactIPMap[pContact.ip] = true
	}
}

func (s *Search) calcSearchTolerance(pContact *Contact) uint32 {
	distance := s.targetID.getXor(pContact.getKadID())
	return distance.get32BitChunk(0)
}

func (s *Search) addFiles(files []*Ed2kFileStruct) []*Ed2kFileStruct {
	var newFiles []*Ed2kFileStruct
	for _, file := range files {
		if s.fileHashMap[file.Hash] {
			continue
		}

		//log.Println(file.Name)

		// check if yellow media or not
		if com.IsYellow(file.Name) {
			continue
		}

		s.fileHashMap[file.Hash] = true

		s.files = append(s.files, file)
		newFiles = append(newFiles, file)
	}

	return newFiles
}

// Conver to file link according to user search keywords
func (s *Search) convert2FileLink(file *Ed2kFileStruct) *com.Ed2kFileLink {
	if file.Type != "Video" {
		return nil
	}

	// filtered by matched items
	fileInfo := com.ToFileInfo(file.Name, s.myKeywordStruct.Items)
	if fileInfo == nil {
		return nil
	}

	// check if season matched with user input
	// we don't care about episode
	if s.myKeywordStruct.MyKeyword.Season != -1 && s.myKeywordStruct.MyKeyword.Season != fileInfo.Season {
		return nil
	}

	fileLink := com.Ed2kFileLink{FileInfo: *fileInfo, Name: file.Name, Size: file.Size, Avail: file.Avail, Hash: file.Hash[:]}

	return &fileLink
}

// Conver to file links according to user search keywords
func (s *Search) convert2FileLinks(files []*Ed2kFileStruct) []*com.Ed2kFileLink {
	var fileLinks []*com.Ed2kFileLink
	for _, file := range files {
		link := s.convert2FileLink(file)
		if link != nil {
			fileLinks = append(fileLinks, link)
		}
	}

	return fileLinks
}
