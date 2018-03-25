package web

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	hotSearchPeriod = 7 // days
	hotSearchNbr    = 12

	latestSearchNbr = 12
)

// DayStats x
type DayStats struct {
	visitIPCount       int
	searchCount        int
	validSearchCount   int
	successSearchCount int
}

// Stats x
type Stats struct {
	DayStats
	SearchStats
}

// SearchStats x
type SearchStats struct {
	LatestSearches []string
	HotSearches    []string
}

// DaySearchStats is search statistics of today
type DaySearchStats struct {
	date string

	visitIPMap map[string]bool

	count        int // search count
	validCount   int
	successCount int
}

func (s *DaySearchStats) init() {
	date := getToday()
	if date != s.date {
		s.date = date

		s.visitIPMap = make(map[string]bool)
		s.count = 0
		s.validCount = 0
		s.successCount = 0
	}
}

func (s *DaySearchStats) add(ip string) {
	s.init()

	s.visitIPMap[ip] = true
	s.count++
}

func (s *DaySearchStats) addValid() {
	s.validCount++
}

func (s *DaySearchStats) addSuccess() {
	s.successCount++
}

// DayHotSearch x
type DayHotSearch struct {
	date       string
	keywordMap map[string]int
}

// HotSearchStats is hot search statistics
type HotSearchStats struct {
	// sorted hot searches
	nbr              int                  // search number
	searches         [hotSearchNbr]string // head is hotest
	searchKeywordMap map[string]bool

	keywordMap  map[string]int // total keyword map
	daySearches []*DayHotSearch
}

func (s *HotSearchStats) Len() int {
	return s.nbr
}

func (s *HotSearchStats) Swap(i, j int) {
	s.searches[i], s.searches[j] = s.searches[j], s.searches[i]
}

func (s *HotSearchStats) Less(i, j int) bool {
	return s.keywordMap[s.searches[i]] > s.keywordMap[s.searches[j]]
}

func (s *HotSearchStats) addSortExclude(key string) {
	// insert sort algorithm
	i := s.nbr - 1
	count := s.keywordMap[key]
	for ; i >= 0; i-- {
		if count > s.keywordMap[s.searches[i]] {
			if i < hotSearchNbr-1 {
				s.searches[i+1] = s.searches[i] // move back
			} else {
				s.nbr--
				delete(s.searchKeywordMap, s.searches[i])
			}
		} else {
			break
		}
	}

	if i < hotSearchNbr-1 { // insert
		s.searches[i+1] = key
		s.searchKeywordMap[key] = true
		s.nbr++
	}
}

func (s *HotSearchStats) addSortInclude() {
	sort.Sort(s)
}

func (s *HotSearchStats) sort() {
	s.nbr = 0 // new sort

	for key := range s.keywordMap {
		s.addSortExclude(key)
	}
}

func (s *HotSearchStats) add(keywords []string) {
	date := getToday()

	// get day search
	if len(s.daySearches) == 0 {
		s.daySearches = append(s.daySearches, &DayHotSearch{date: date, keywordMap: make(map[string]int)})
	}

	dayHotSearch := s.daySearches[len(s.daySearches)-1]
	if dayHotSearch.date != date {
		dayHotSearch = &DayHotSearch{date: date, keywordMap: make(map[string]int)}
		s.daySearches = append(s.daySearches, dayHotSearch)
	}

	// increase keyword count
	for _, key := range keywords {
		dayHotSearch.keywordMap[key]++
		s.keywordMap[key]++
	}

	if len(s.daySearches) > hotSearchPeriod {
		// remove keyword counts of past day
		for key, count := range s.daySearches[0].keywordMap {
			s.keywordMap[key] -= count
		}

		// re-sort again
		s.daySearches = s.daySearches[1:]
		s.sort()
	} else {
		for _, key := range keywords {
			if s.searchKeywordMap[key] {
				s.addSortInclude()
			} else {
				s.addSortExclude(key)
			}
		}
	}
}

// UserSearchTrack is for tracking user searches.
/////////////////////////////////////////////////
type UserSearchTrack struct {
	daySearchStats DaySearchStats
	hotSearchStats HotSearchStats

	// latest search
	latestSearches []string // tail is latest

	lock sync.Mutex
}

func (u *UserSearchTrack) getLatestSearches() []string {
	// reverse
	nbr := len(u.latestSearches)
	latestSearches := make([]string, nbr)
	for i := 0; i < nbr; i++ {
		latestSearches[nbr-i-1] = u.latestSearches[i]
	}

	return latestSearches
}

// NewUserSearchTrack x
func NewUserSearchTrack() *UserSearchTrack {
	return &UserSearchTrack{
		daySearchStats: DaySearchStats{
			visitIPMap: make(map[string]bool),
		},
		hotSearchStats: HotSearchStats{
			searchKeywordMap: make(map[string]bool),
			keywordMap:       make(map[string]int),
		},
	}
}

func (u *UserSearchTrack) addLatestSearch(keywords []string) {
	s := u.latestSearches
	for _, key := range keywords {
		existing := false
		for i, latest := range s {
			if key == latest {
				s = append(s[:i], s[i+1:]...)
				s = append(s, key)
				existing = true
				break
			}
		}

		if !existing {
			s = append(s, key)
		}
	}

	if len(s) > latestSearchNbr {
		s = s[1:]
	}

	u.latestSearches = s
}

func (u *UserSearchTrack) addHotSearch(keywords []string) {
	u.hotSearchStats.add(keywords)
}

func (u *UserSearchTrack) addSearchKeywords(keywords []string) {
	u.lock.Lock()
	u.addLatestSearch(keywords)
	u.addHotSearch(keywords)
	u.lock.Unlock()
}

func (u *UserSearchTrack) visit() {
	u.lock.Lock()
	u.daySearchStats.init()
	u.lock.Unlock()
}

func (u *UserSearchTrack) addSearch(ip string) {
	u.lock.Lock()
	u.daySearchStats.add(ip)
	u.lock.Unlock()
}

func (u *UserSearchTrack) addValidSearch() {
	u.lock.Lock()
	u.daySearchStats.addValid()
	u.lock.Unlock()
}

func (u *UserSearchTrack) addSuccessSearch() {
	u.lock.Lock()
	u.daySearchStats.addSuccess()
	u.lock.Unlock()
}

func (u *UserSearchTrack) getStats() *Stats {
	u.lock.Lock()

	status := &Stats{
		DayStats: DayStats{visitIPCount: len(u.daySearchStats.visitIPMap),
			searchCount:        u.daySearchStats.count,
			validSearchCount:   u.daySearchStats.validCount,
			successSearchCount: u.daySearchStats.successCount,
		},

		SearchStats: SearchStats{LatestSearches: u.getLatestSearches(),
			HotSearches: u.hotSearchStats.searches[:u.hotSearchStats.nbr],
		},
	}

	u.lock.Unlock()

	return status
}

func (u *UserSearchTrack) getSearchStats() *SearchStats {
	u.lock.Lock()

	stats := &SearchStats{LatestSearches: u.getLatestSearches(),
		HotSearches: u.hotSearchStats.searches[:u.hotSearchStats.nbr],
	}

	u.lock.Unlock()

	return stats
}

func getToday() string {
	now := time.Now()

	year := now.Year()
	month := now.Month()
	day := now.Day()
	return fmt.Sprintf("%d-%d-%d", year, month, day)
}
