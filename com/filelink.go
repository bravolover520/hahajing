package com

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	digits = "0123456789"
)

// FileInfo x
type FileInfo struct {
	Type    byte
	OrgName string
	ChName  string
	Season  int
	Episode int
}

// Ed2kFileLink x
type Ed2kFileLink struct {
	FileInfo

	// attributes from ED2K
	Name  string
	Size  uint64
	Avail uint32
	Hash  []byte
}

type ed2kFileLinkJSON struct {
	FileInfo

	// attributes from ED2K
	Name  string
	Size  uint64
	Avail uint32
	Link  string
}

// @name: lower case
// @orgName: lower case
func parseName(name string, orgName string) bool {
	// check if containing orginal name
	keywords := Split2Keywords(orgName)
	pattern := strings.Join(keywords, ".*")
	match, _ := regexp.MatchString(pattern, name)
	return match
}

// @name: lower case
// @orgName: lower case
func parseSeasonTVName(name string, orgName string) (int, int) {
	match := parseName(name, orgName)
	if !match {
		return -1, -1
	}

	return getSeasonEpisode(name)
}

// @name: lower case
// @orgName: lower case
func parseNoSeasonTVName(name string, orgName string) int {
	match := parseName(name, orgName)
	if !match {
		return -1
	}

	return getEpisode(name)
}

// @name: lower case
// @orgName: lower case
func parseUnknownTypeName(name string, orgName string) (int, int, byte) {
	match := parseName(name, orgName)
	if !match {
		return -1, -1, UnknownType
	}

	season, episode := getSeasonEpisode(name)
	if season != -1 && episode != -1 {
		return season, episode, SeasonTV
	}

	episode = getEpisode(name)
	if episode != -1 {
		return -1, episode, NoSeasonTV
	}

	// movie
	return -1, -1, Movie
}

// Get episode without seasons
// like ep01, e01, .01., 1x01, -01
func getEpisode(name string) int {
	episode := -1
	state := 0
	dot := false
	for _, c := range name {
		s := string(c)
		switch state {
		case 0: // start
			if s == "e" || s == "." || s == "x" {
				if s == "." {
					dot = true
				}
				state = 1
			}
		case 1: // episode
			i := strings.Index(digits, s)
			if i != -1 {
				if episode == -1 {
					episode = i
				} else {
					episode = episode*10 + i
				}
			} else {
				if episode != -1 && episode < 200 {
					if !dot || (dot && s == ".") {
						return episode
					}
				}

				episode = -1

				if s != "p" && s != "." && s != "e" && s != "x" {
					dot = false
					state = 0
				}

				if s == "." {
					dot = true
				}
			}
		}
	}

	if episode != -1 && episode < 200 {
		return episode
	}

	return -1
}

func getSeasonEpisode(name string) (int, int) {
	season, episode := -1, -1
	state := 0
	for _, c := range name {
		s := string(c)
		switch state {
		case 0: // start
			i := strings.Index(digits, s)
			if i != -1 {
				season = i
				state = 1
			} else if s == "s" {
				state = 1
			}
		case 1: // season
			i := strings.Index(digits, s)
			if i != -1 {
				if season == -1 {
					season = i
				} else {
					season = season*10 + i
				}
			} else {
				if season < 0 || season > 100 { // ignore like 1024x768 and make sure season is reasonable.
					season = -1
					state = 0
				} else {
					if s == "e" || s == "x" {
						state = 2
					} else {
						state = 3
					}
				}
			}
		case 2: // episode
			i := strings.Index(digits, s)
			if i != -1 {
				if episode == -1 {
					episode = i
				} else {
					episode = episode*10 + i
				}
			} else {
				if episode == -1 {
					season = -1
					state = 0
				} else {
					return season, episode
				}
			}
		case 3: // pause, like s01.e01
			if s == "e" {
				state = 2
			} else {
				i := strings.Index(digits, s)
				if i == -1 {
					season = -1
					state = 0
				} else {
					season = i
					state = 1
				}
			}
		}
	}

	return season, episode
}

// ToFileInfo is converting file name from KAD or DHT via item from Internet(DouBan)
// @items: already sorted
func ToFileInfo(name string, items []*Item) *FileInfo {
	// match
	lowerName := strings.ToLower(name)
	var fileInfo *FileInfo
	for _, item := range items {
		if fileInfo != nil {
			break
		}

		orgName := strings.ToLower(item.OrgName)

		switch item.Type {
		case SeasonTV:
			season, episode := parseSeasonTVName(lowerName, orgName)
			if season != -1 && episode != -1 {
				fileInfo = &FileInfo{Type: item.Type, OrgName: item.OrgName, ChName: item.ChName,
					Season:  season,
					Episode: episode}
			}
		case NoSeasonTV:
			episode := parseNoSeasonTVName(lowerName, orgName)
			if episode != -1 {
				fileInfo = &FileInfo{Type: item.Type, OrgName: item.OrgName, ChName: item.ChName,
					Season:  -1,
					Episode: episode}
			}
		case Movie:
			match := parseName(lowerName, orgName)
			if match {
				fileInfo = &FileInfo{Type: item.Type, OrgName: item.OrgName, ChName: item.ChName,
					Season:  -1,
					Episode: -1}
			}
		default:
			season, episode, itemType := parseUnknownTypeName(lowerName, orgName)
			if itemType != UnknownType {
				// Note that we don't set type of item because item type cannot be inferred by movie/tv name.
				fileInfo = &FileInfo{Type: itemType, OrgName: item.OrgName, ChName: item.ChName,
					Season:  season,
					Episode: episode}
			}
		}
	}

	if fileInfo == nil {
		return nil
	}

	// accurate check for yellow check
	if fileInfo.OrgName != fileInfo.ChName {
		// Not Chinese movie or TV
		// If file name have Chinese char, we think it can be matched by its Chinese name.
		// Otherwise, we think it as AV.
		hasCh := false
		for _, c := range lowerName {
			if IsChinese(c) {
				hasCh = true
				break
			}
		}

		if hasCh {
			chName := strings.ToLower(fileInfo.ChName)
			if strings.Index(lowerName, chName) == -1 {
				return nil
			}
		}
	}

	return fileInfo
}

// GetEd2kLink x
func (f *Ed2kFileLink) GetEd2kLink() string {
	return GetEd2kLink(f.Name, f.Size, f.Hash)
}

// GetHash x
func (f *Ed2kFileLink) GetHash() [16]byte {
	var hashArray [16]byte
	copy(hashArray[:], f.Hash)
	return hashArray
}

// GetPrintStr x
func (f *Ed2kFileLink) GetPrintStr() string {
	log.Printf("Name: %s, Size: %d, Avail:%d\nEd2k: %s\n", f.Name, f.Size, f.Avail, f.GetEd2kLink())

	return fmt.Sprintf("Name: %s, Size: %d, Avail:%d\nEd2k: %s\n", f.Name, f.Size, f.Avail, f.GetEd2kLink())
}

// ToJSON x
func (f *Ed2kFileLink) ToJSON() []byte {
	linkJSON := ed2kFileLinkJSON{
		FileInfo: f.FileInfo,
		Name:     f.Name,
		Size:     f.Size,
		Avail:    f.Avail,
		Link:     f.GetEd2kLink()}

	b, _ := json.Marshal(linkJSON)
	return b
}

// GetEd2kLink is getting ED2K link by file name, size and hash from eMule KAD network.
func GetEd2kLink(name string, size uint64, hash []byte) string {
	newHash := ConvertEd2kHash32(hash)
	return fmt.Sprintf("ed2k://|file|%s|%d|%s|/",
		encodeURLUtf8(stripInvalidFileNameChars(name)),
		size,
		encodeBase16(newHash[:]))
}

// ConvertEd2kHash32 x
func ConvertEd2kHash32(srcHash []byte) [16]byte {
	// change to inverse endian for each uint32
	hash := [16]byte{}
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			hash[i*4+j] = srcHash[i*4+3-j]
		}
	}
	return hash
}

var reservedFileNames = [...]string{"NUL", "CON", "PRN", "AUX", "CLOCK$",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}

func stripInvalidFileNameChars(text string) string {
	var dst string
	for _, c := range text {
		if (c >= 0 && c <= 31) ||
			c == '"' || c == '*' || c == '<' || c == '>' ||
			c == '?' || c == '|' || c == '\\' || c == ':' {
			continue
		}
		dst += string(c)
	}

	for _, prefix := range reservedFileNames {
		if len(dst) < len(prefix) {
			continue
		}

		if dst[:len(prefix)] == prefix {
			if len(dst) == len(prefix) {
				dst += string('_')
			} else if dst[len(prefix)] == '.' {
				s := []rune(dst)
				s[len(prefix)] = '_'
				dst = string(s)
			}
		}
	}

	return dst
}

/*
func encodeURLUtf8(str string) string {
	return url.PathEscape(str)
}
*/

func encodeURLUtf8(str string) string {
	utf8 := []byte(str)
	var url string
	for _, b := range utf8 {
		if b == byte('%') || b == byte(' ') || b >= 0x7F {
			s := fmt.Sprintf("%%%02X", b)
			url += s
		} else {
			url += string(b)
		}
	}

	return url
}

func encodeBase16(buf []byte) string {
	s := hex.EncodeToString(buf)
	return strings.ToUpper(s)
}
