package web

import (
	"encoding/json"
	"fmt"
	"hahajing/com"
	"hahajing/door"
	"hahajing/kad"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/websocket"
)

const (
	keywordCheckWaitingTime = 5
	kadSearchWaitingTime    = 5
)

// webError is for user browser.
type webError struct {
	Error string
}

// Web x
type Web struct {
	searchReqCh       chan *kad.SearchReq
	keywordCheckReqCh chan *door.KeywordCheckReq

	homeTemplate    *template.Template
	keywordManager  *com.KeywordManager
	userSearchTrack *UserSearchTrack
}

// Start x
func (we *Web) Start(searchReqCh chan *kad.SearchReq, keywordCheckReqCh chan *door.KeywordCheckReq, keywordManager *com.KeywordManager) {
	we.searchReqCh = searchReqCh
	we.keywordCheckReqCh = keywordCheckReqCh
	we.keywordManager = keywordManager

	we.userSearchTrack = NewUserSearchTrack()

	// HTML page
	path := com.GetConfigPath()
	tmpl, err := template.ParseFiles(path + "/config/web/home.html")
	if err != nil {
		log.Panic("Home page failed!")
	}
	we.homeTemplate = tmpl

	// at last start sever
	we.startServer()
}

func (we *Web) checkKeywordsFromDoor(myKeyword *com.MyKeyword) ([]*com.Item, string) {
	// send Door to check if this keyword is legal
	resCh := make(chan *door.KeywordCheckRes, 1)
	req := door.KeywordCheckReq{ResCh: resCh, MyKeyword: myKeyword}
	we.keywordCheckReqCh <- &req

	// waiting result from Door
	select {
	case res := <-resCh:
		return res.Items, res.ErrorStr
	case <-time.After(keywordCheckWaitingTime * time.Second):
		return nil, "检查超时，请重试！"
	}
}

func (we *Web) readSearchInput(ws *websocket.Conn) (*com.MyKeyword, string) {
	// read
	msg := make([]byte, 1024)
	n, err := ws.Read(msg)
	if err != nil {
		return nil, "读取数据错误，请重试！"
	}

	// add new search
	ip, _, _ := net.SplitHostPort(ws.Request().RemoteAddr)
	we.userSearchTrack.addSearch(ip)

	// parse
	text := string(msg[:n])
	com.HhjLog.Infof("New user(%s) input: %s", ip, text)

	keywords := com.Split2Keywords(text)
	if keywords == nil {
		return nil, "没有搜索关键字，请重新输入！"
	}

	myKeyword := com.NewMyKeyword(keywords)
	if len(myKeyword.SearchKeywords) == 0 {
		return nil, "无有效搜索关键字，请重新输入！"
	}

	return myKeyword, ""
}

func (we *Web) checkKeywordsFromKeywordManager(myKeyword *com.MyKeyword) ([]*com.Item, bool) {
	// get from keyword manager
	items := we.keywordManager.Get(myKeyword.SearchKeywords)
	if items == nil { // not existing in keyword manager
		return nil, false
	}

	// filter
	return com.FilterItems(items, myKeyword), true
}

func (we *Web) writeError(ws *websocket.Conn, errStr string) {
	data, _ := json.Marshal(&webError{Error: errStr})
	ws.Write(data)
}

func (we *Web) send2Kad(ws *websocket.Conn, myKeywordStruct *com.MyKeywordStruct) {
	resCh := make(chan *kad.SearchRes, kad.SearchResChSize)
	searchReq := kad.SearchReq{ResCh: resCh, MyKeywordStruct: myKeywordStruct}
	we.searchReqCh <- &searchReq

	// waiting result from KAD
	found := false
	fileLinks := make(map[[16]byte]bool)
	for {
		select {
		case pSearchRes := <-resCh:
			for _, fileLink := range pSearchRes.FileLinks {
				if !found {
					we.userSearchTrack.addSuccessSearch()
				}
				found = true
				hash := fileLink.GetHash()
				if !fileLinks[hash] {
					ws.Write(fileLink.ToJSON())
				}
			}
		case <-time.After(kadSearchWaitingTime * time.Second):
			if !found {
				we.writeError(ws, "搜索超时，请重试！")
			}
			return
		}
	}
}

func (we *Web) searchHandler(ws *websocket.Conn) {
	// read user input from network
	myKeyword, errStr := we.readSearchInput(ws)
	if myKeyword == nil {
		we.writeError(ws, errStr)
		return
	}

	// firstly, check from keyword manager
	items, bInDataBase := we.checkKeywordsFromKeywordManager(myKeyword)
	if !bInDataBase {
		// send to Door for keyword checking
		items, errStr = we.checkKeywordsFromDoor(myKeyword)
	}
	if len(items) == 0 {
		if errStr == "" {
			errStr = "没有找到相关的电视剧或者电影，请重新输入！"
		}
		we.writeError(ws, errStr)
		return
	}

	// add valid new search keywords for statistic
	we.userSearchTrack.addSearchKeywords([]string{strings.Join(myKeyword.SearchKeywords, " ")})
	we.userSearchTrack.addValidSearch()

	// send to KAD
	myKeywordStruct := com.NewMyKeywordStruct(myKeyword, items)
	if myKeywordStruct == nil {
		we.writeError(ws, "关键字错误，请重新输入！")
		return
	}
	we.send2Kad(ws, myKeywordStruct)
}

func (we *Web) homeHandler(w http.ResponseWriter, r *http.Request) {
	homeData := &HomeData{Host: "ws://" + r.Host + "/search",
		SearchStats: we.userSearchTrack.getSearchStats(),
	}
	err := we.homeTemplate.Execute(w, homeData)
	if err != nil {
		com.HhjLog.Criticalf("Execute template failed: %s", err)
	}

	we.userSearchTrack.visit()
}

func (we *Web) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := we.userSearchTrack.getStats()
	s := fmt.Sprintf("今日IP访问: %d, 搜索: %d, 有效搜索: %d, 成功搜索: %d", stats.visitIPCount, stats.searchCount, stats.validSearchCount, stats.successSearchCount)

	w.Write([]byte(s))
}

func (we *Web) startServer() {
	com.HhjLog.Info("Web Server is running...")

	http.HandleFunc("/", we.homeHandler)
	http.HandleFunc("/1979", we.statsHandler)
	http.Handle("/search", websocket.Handler(we.searchHandler))

	var err error
	if len(os.Args) > 1 && os.Args[1] == "server" {
		err = http.ListenAndServe(":80", nil)
	} else {
		err = http.ListenAndServe(":66", nil)
	}
	if err != nil {
		com.HhjLog.Panic("Start Web Server failed: ", err)
	}
}
