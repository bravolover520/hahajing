package com

import (
	"os"
	"strings"

	"github.com/op/go-logging"
)

// HhjLog is HHJ system log
var HhjLog = logging.MustGetLogger("hhj")
var logformat = logging.MustStringFormatter(
	`%{color}%{time:2006-01-02 15:04:05.000} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

func init() {
	backend := logging.NewLogBackend(os.Stdout, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, logformat)

	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.CRITICAL, "")

	// Set the backends to be used.
	logging.SetBackend(backendLeveled, backendFormatter)
}

// GetConfigPath x
func GetConfigPath() string {
	path := os.Args[0]

	i := strings.LastIndex(path, "\\")
	if i == -1 {
		i = strings.LastIndex(path, "/")
	}

	if i == -1 {
		HhjLog.Fatalf("Config path error: %s", path)
	}

	path = string(path[0:i])

	return path
}

// CreatePath x
func CreatePath(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	// create new path
	if os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	}

	return err
}

// CreateFile x
func CreateFile(fileName string) error {
	// create file if not existing
	_, err := os.Stat(fileName)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		f, err := os.Create(fileName)
		if err != nil {
			return err
		}

		f.Close()
		return nil
	}

	return err
}

// StripString is for string cannot be legal file name or directory name at Windows. Not sure if it's same rule at Linux.
func StripString(s string) string {
	for _, c := range `[/\:*?"<>|]` {
		s = strings.Replace(s, string(c), "", -1)
	}

	return s
}

// IsChinese x
func IsChinese(c rune) bool {
	// 标点符号
	for _, c1 := range `‘’“”；、：，。？！` {
		if c1 == c {
			return true
		}
	}

	// 基本汉字	20902字	4E00-9FA5
	// rune is unicode code point
	return c >= 0x4E00 && c <= 0x9FA5
}

// IsEnglishOrChinese is char is English or Chinese.
func IsEnglishOrChinese(c rune) bool {
	if c >= 0 && c <= 255 {
		return true
	}

	return IsChinese(c)
}

// Split2PrimaryKeywords is split to slice of primary keyword by seperators.
// And keyword containing specific char not thinking as primary keyword.
func Split2PrimaryKeywords(s string) []string {
	ignore := "'’"
	keys := innerSplit2Keywords(s, ignore)

	var newKeys []string
	for _, key := range keys {
		valid := true
		for _, c := range ignore {
			if strings.Index(key, string(c)) != -1 {
				valid = false
				break
			}
		}

		if valid {
			newKeys = append(newKeys, key)
		}
	}

	return newKeys
}

// Split2Keywords is split to slice of keyword by seperators.
func Split2Keywords(s string) []string {
	return innerSplit2Keywords(s, "")
}

// innerSplit2Keywords is split to slice of keyword by seperators.
// @ignore: which chars not think as seperator.
func innerSplit2Keywords(s string, ignore string) []string {
	sep := `·!/\*?<>|-_:,.;'"()[]‘’“”；、：，。？！` + "\t"
	for _, c := range ignore {
		sep = strings.Replace(sep, string(c), "", -1)
	}

	for _, c := range sep {
		s = strings.Replace(s, string(c), " ", -1)
	}

	s = strings.ToLower(s)

	var newKeys []string
	for _, key := range strings.Split(s, " ") {
		if key != "" {
			newKeys = append(newKeys, key)
		}
	}

	return newKeys
}

var yellowKeys = []string{
	"性交", "做爱", "打炮", "无码", "有码", "淫", "偷拍", "中出", "熟女", "巨乳", "人妻",
	"無碼", "有碼",
	"sex", "gay",
}

var yellowGroupKeys = [][]string{
	{"tokyo", "hot"},
}

// IsYellow is to check if name has sex info or not.
func IsYellow(name string) bool {
	name = strings.ToLower(name)

	// filter Japanse
	for _, c := range name {
		if !IsEnglishOrChinese(c) {
			return true
		}
	}

	// filter by keyword
	for _, key := range yellowKeys {
		if strings.Index(name, key) != -1 {
			return true
		}
	}

	// filter by group keywords
	for _, keys := range yellowGroupKeys {
		yellow := true
		for _, key := range keys {
			if strings.Index(name, key) == -1 {
				yellow = false
				break
			}
		}

		if yellow {
			return true
		}
	}

	return false
}
