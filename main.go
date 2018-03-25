package main

import (
	"hahajing/com"
	"hahajing/door"
	"hahajing/kad"
	"hahajing/web"
)

var kadInstance kad.Kad
var webInstance web.Web
var doorInstance door.Door
var keywordManager = com.NewKeywordManager()

func main() {
	kadInstance.Start()
	doorInstance.Start(keywordManager)

	webInstance.Start(kadInstance.SearchReqCh, doorInstance.KeywordCheckReqCh, keywordManager)
}
