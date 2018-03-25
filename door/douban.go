package door

import (
	"hahajing/com"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const maxDouBanTryNbr = 3

// DouBan is crawler. Cocurrent isn't supported.
type DouBan struct {
	client *http.Client
}

func (db *DouBan) start() bool {
	// Get cookie firstly
	jar, _ := cookiejar.New(nil)
	db.client = &http.Client{Jar: jar}

	req, _ := http.NewRequest("GET", "https://www.douban.com/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")

	res, err := db.client.Do(req)
	if err != nil {
		com.HhjLog.Warningf("New DouBan session failed: %s", err)
		return false
	}
	res.Body.Close()

	return true
}

func (db *DouBan) search(keywords []string, level int) []*com.Item {
	params := "q=" + url.QueryEscape(strings.Join(keywords, " "))
	url := "https://www.douban.com/search?cat=1002&" + params
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")

	res, err := db.client.Do(req)
	if err != nil || res.StatusCode != 200 {
		level++
		if level > maxDouBanTryNbr {
			com.HhjLog.Warning("Reach max DouBan retries!")
			return nil
		}

		// new session
		if !db.start() {
			com.HhjLog.Warning("Start DouBan failed during search!")
			return nil
		}

		// search it again
		return db.search(keywords, level)
	}

	return db.getItems(res)
}

func (db *DouBan) getItems(res *http.Response) []*com.Item {
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil
	}

	var items []*com.Item
	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		sTemp := s.Find("h3 span")
		text := sTemp.First().Text()

		// type
		tv := true
		if strings.Index(text, "电影") != -1 {
			tv = false
		} else if strings.Index(text, "电视剧") == -1 {
			return
		}

		// Chinese name
		sTemp = s.Find("h3 a")
		text = strings.TrimSpace(sTemp.Text())
		chNames1 := strings.Split(text, " ")
		var chNames []string
		for _, name := range chNames1 {
			if name != "" {
				chNames = append(chNames, name)
			}
		}

		// check if it is Chinese name or not
		// if not, just ignore this item
		chName := chNames[0]
		if len(chName) == len([]rune(chName)) {
			return
		}

		// get type by Chinese name
		byType := com.Movie
		if tv {
			byType = com.NoSeasonTV
			if len(chNames) > 1 { // check if has season
				chSeason := []rune(chNames[1])
				if chSeason[0] == '第' && chSeason[len(chSeason)-1] == '季' {
					byType = com.SeasonTV
				}
			}
		}

		// original name
		sTemp = s.Find(".subject-cast")
		text = sTemp.Text()
		texts := strings.Split(text, " /")
		text = texts[0]
		orgName := string([]rune(text)[3:])
		orgName = strings.TrimSpace(orgName)

		// only support Chinese or English orginal name
		for _, c := range []rune(orgName) {
			if !com.IsEnglishOrChinese(c) {
				return
			}
		}

		// new item
		item := com.Item{Type: byType, OrgName: orgName, ChName: chName}

		// add different item
		for _, pItem := range items {
			if item == *pItem {
				return
			}
		}

		com.HhjLog.Infof("New item from DouBan: %+v", item)
		items = append(items, &item)
	})

	return items
}
