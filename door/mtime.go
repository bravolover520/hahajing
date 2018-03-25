package door

import (
	"encoding/json"
	"fmt"
	"hahajing/com"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const mtimeURL = "http://service.channel.mtime.com/Search.api?Ajax_CallBack=true&Ajax_CallBackType=Mtime.Channel.Services&Ajax_CallBackMethod=GetSearchResult&Ajax_CallBackArgument0=%s&Ajax_CallBackArgument1=1&Ajax_CallBackArgument2=290&Ajax_CallBackArgument3=0&Ajax_CallBackArgument4=1"

// MTime x
type MTime struct {
}

func (m *MTime) newItem(title, otherTitle string, mediaLength int) *com.Item {
	i := strings.Index(title, " ")
	if i == -1 {
		return nil
	}
	chName := title[:i]

	j := strings.Index(title, " (")
	if j == -1 {
		return nil
	}

	orgName := chName
	if i < j {
		orgName = title[i+1 : j]
	}

	// other Chinese name
	otherChName := ""
	i = strings.Index(otherTitle, "更多名：")
	if i != -1 {
		otherChName = otherTitle[i+len("更多名："):]
	}

	byType := com.UnknownType
	if mediaLength >= 120 { // 120 minutes
		byType = com.Movie
	}

	return &com.Item{Type: byType, OrgName: orgName, ChName: chName, OtherChName: otherChName}
}

func (m *MTime) getItems(data []byte) []*com.Item {
	if len(data) < len("var getSearchResult = ")+3 {
		return nil
	}

	// parse
	var f interface{}
	jsonData := data[len("var getSearchResult = ") : len(data)-3]
	err := json.Unmarshal([]byte(jsonData), &f)
	if err != nil {
		return nil
	}

	v, ok := f.(map[string]interface{})
	if !ok {
		return nil
	}

	v, ok = v["value"].(map[string]interface{})
	if !ok {
		return nil
	}

	v, ok = v["movieResult"].(map[string]interface{})
	if !ok {
		return nil
	}

	movies, ok := v["moreMovies"].([]interface{})
	if !ok {
		return nil
	}

	// get items
	var items []*com.Item
	for _, mv := range movies {
		item, ok := mv.(map[string]interface{})
		if !ok {
			return nil
		}

		title, ok := item["movieTitle"].(string)
		if !ok {
			return nil
		}

		otherTitle, ok := item["titleOthers"].(string)
		mediaLength, ok := item["movieLength"].(float64)
		pItem := m.newItem(title, otherTitle, int(mediaLength))
		if pItem != nil {
			com.HhjLog.Infof("New item from MTime: %+v", *pItem)
			items = append(items, pItem)
		}
	}

	return items
}

func (m *MTime) search(keywords []string) []*com.Item {
	client := &http.Client{}

	params := url.QueryEscape(strings.Join(keywords, " "))
	url := fmt.Sprintf(mtimeURL, params)
	res, err := client.Get(url)
	if err != nil || res.StatusCode != 200 {
		return nil
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}

	items := m.getItems(data)

	res.Body.Close()

	return items
}
