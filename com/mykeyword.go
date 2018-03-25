package com

import (
	"strconv"
	"strings"
)

const (
	// Movie type
	Movie byte = 0

	// SeasonTV is TV has season
	SeasonTV byte = 1

	// NoSeasonTV is TV has no season
	NoSeasonTV byte = 2

	// UnknownType x
	UnknownType byte = 255
)

const (
	minTargetKeywordSize  = 3
	minPrimaryKeywordSize = 3
)

// MyKeyword is keyword in my system, low cases except @OrgKeywords
type MyKeyword struct {
	OrgKeywords []string // from user input

	SearchKeywords []string // search keywords of name, for DouBan search
	NameKeywords   []string // keywords of name based on user input
	Season         int      // 0: all seasons for TV and movies
}

// Item is getting from internet, e.g. DouBan
type Item struct {
	Type        byte
	OrgName     string
	ChName      string
	OtherChName string
}

// MyKeywordStruct is used for KAD search with multiple target keywords.
type MyKeywordStruct struct {
	TargetKeywords []string // For KAD search

	MyKeyword *MyKeyword // from user
	Items     []*Item    // from Internet or database
}

var theWords = map[string]bool{"the": true, "these": true, "that": true, "a": true, "this": true,
	"he": true, "she": true, "we": true, "you": true, "us": true, "his": true, "her": true, "it": true, "my": true, "our": true,
	"no": true, "yes": true, "not": true, "is": true, "are": true,
	"in": true, "on": true, "of": true}

var theChDigits = map[string]int{"零": 0, "一": 1, "二": 2, "三": 3, "四": 4, "五": 5, "六": 6, "七": 7, "八": 8, "九": 9}

func parseDigit(s string, bCh bool) (int, bool) {
	// @s is keyword, which means no space or other else.
	// For Chinese parse, only identify less than 100.
	v, err := strconv.Atoi(s)
	if err == nil {
		return v, true
	}

	if !bCh {
		return -1, false
	}

	// think it as Chinese
	v = -1
	for _, c := range s {
		if c == '十' {
			if v == -1 {
				v = 1
			}
			v *= 10
		} else {
			d, ok := theChDigits[string(c)]
			if ok {
				if v == -1 {
					v = 0
				}
				v += d
			} else {
				return -1, false
			}
		}
	}

	return v, true
}

// NewMyKeyword is converting keywords from user to my format
func NewMyKeyword(keywords []string) *MyKeyword {
	var ignoreI = -1
	myKeyword := MyKeyword{Season: -1}
	myKeyword.OrgKeywords = keywords // keywords from user

	for i, key := range keywords {
		if i == ignoreI {
			continue
		}

		// check if season or episode keyword.
		// if so, only extract season or episode information, without thinking it as keyword.
		if key == "season" {
			// assume next keyword is specific season
			if i < len(keywords)-1 {
				season, ok := parseDigit(keywords[i+1], true)
				if ok {
					myKeyword.Season = season
					ignoreI = i + 1
					continue
				}
			}
		} else {
			// get digit string
			text := []rune(key)
			var seasonStr string
			switch text[0] {
			case 's':
				seasonStr = string(text[1:])
			case '第':
				if text[len(text)-1] == '集' { // skip like 第三集
					continue
				}

				// like 第三季, 第三
				end := len(text)
				if text[len(text)-1] == '季' {
					end = len(text) - 1
				}
				seasonStr = string(text[1:end])
			}

			// convert to int
			if seasonStr != "" {
				season, ok := parseDigit(seasonStr, true)
				if ok {
					myKeyword.Season = season
					continue
				}
			}
		}

		// name keyword
		myKeyword.NameKeywords = append(myKeyword.NameKeywords, key)

		// search keyword
		// used for DouBan search
		myKeyword.SearchKeywords = append(myKeyword.SearchKeywords, GetPrimaryKeywordsByKeyword(key)...)
	}

	return &myKeyword
}

// FilterItems is checking items from internet/database if satisfying user search keyword.
// We use NameKeywords, not PrimaryKeywords for accurate matching.
func FilterItems(m []*Item, myKeyword *MyKeyword) []*Item {
	// filter
	var items []*Item
	for _, item := range m {
		orgName := strings.ToLower(item.OrgName)
		bSatisfied := true
		for _, key := range myKeyword.NameKeywords {
			if strings.Index(orgName, key) == -1 &&
				strings.Index(item.ChName, key) == -1 &&
				strings.Index(item.OtherChName, key) == -1 {
				bSatisfied = false
				break
			}
		}

		if bSatisfied {
			items = append(items, item)
		}
	}

	// sort it for later file name classification
	// TV first
	var tvItems, movieItems, otherItems []*Item
	for _, item := range items {
		switch item.Type {
		case SeasonTV:
			fallthrough
		case NoSeasonTV:
			tvItems = append(tvItems, item)
		case Movie:
			movieItems = append(movieItems, item)
		default: // Unknown type
			otherItems = append(otherItems, item)
		}
	}

	newItems := append(tvItems, movieItems...)
	newItems = append(newItems, otherItems...)

	return newItems
}

// GetPrimaryKeywords is getting primary keyword slice and map via name
// Used for KAD search
func GetPrimaryKeywords(s string) ([]string, map[string]bool) {
	keywordMap := make(map[string]bool)
	var keywordSlice []string

	keys := Split2PrimaryKeywords(s)
	for _, key := range keys {
		newKeys := GetPrimaryKeywordsByKeyword(key)
		for _, newKey := range newKeys {
			if !keywordMap[newKey] {
				keywordSlice = append(keywordSlice, newKey)
				keywordMap[newKey] = true
			}
		}
	}

	if len(keywordSlice) == 0 {
		return nil, nil
	}

	return keywordSlice, keywordMap
}

// GetPrimaryKeywordsByKeyword is get primary keywords by native keyword.
func GetPrimaryKeywordsByKeyword(keyword string) []string {
	if theWords[keyword] {
		return nil
	}

	if len(keyword) < minPrimaryKeywordSize {
		return nil
	}

	return []string{keyword}
}

// GetPrimaryKeywordsByPrimaryKeyword is getting primary keywords via one primary keyword.
// Like 捉妖记2, we think primary keywords are 捉妖记 and 捉妖记2.
// Primary keyword is keyword in syntax.
func GetPrimaryKeywordsByPrimaryKeyword(keyword string) []string {
	keywords := []string{keyword}
	text := []rune(keyword)
	if IsChinese(text[0]) {
		i := len(text) - 1
		for ; i > 0; i-- {
			if text[i] >= '0' && text[i] <= '9' {
				continue
			}
			break
		}

		newKeyword := string(text[:i+1])
		if len(newKeyword) != len(keyword) {
			keywords = append(keywords, newKeyword)
		}
	}

	return keywords
}

// NewMyKeywordStruct is created for KAD search.
func NewMyKeywordStruct(myKeyword *MyKeyword, items []*Item) *MyKeywordStruct {
	targetKeywords := getTargetKeywords(items)
	if targetKeywords == nil {
		HhjLog.Warningf("No target keywords for MyKeyword: %+v", myKeyword)
		return nil
	}

	return &MyKeywordStruct{TargetKeywords: targetKeywords, MyKeyword: myKeyword, Items: items}
}

// get target keywords for KAD
func getTargetKeywords(items []*Item) []string {
	// get target keyword map
	targetKeywordMap := make(map[string]bool)
	for _, item := range items {
		// check if target keyword existing or not
		keywordSlice, keywordMap := GetPrimaryKeywords(item.OrgName)
		existing := false
		for keyword := range keywordMap {
			if targetKeywordMap[keyword] {
				existing = true
				break
			}
		}

		// get new target keyword
		if !existing {
			targetKeyword := getTargetKeyword(keywordSlice, targetKeywordMap)
			if targetKeyword != "" {
				targetKeywordMap[targetKeyword] = true
			}
		}
	}

	// convert to slice
	var targetKeywords []string
	for keyword := range targetKeywordMap {
		targetKeywords = append(targetKeywords, keyword)
	}

	return targetKeywords
}

func getTargetKeyword(primaryKeywords []string, targetKeywordMap map[string]bool) string {
	for _, key := range primaryKeywords {
		if len(key) >= minTargetKeywordSize {
			if !targetKeywordMap[key] {
				return key
			}
		}
	}

	return ""
}
