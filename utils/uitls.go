package utils

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"runtim/debug"
	"runtime"
	"strings"
	"time"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"
	//  Now is a time toolkit for golang
	"github.com/jinzhu/now"
	//	a HTML sanitizer implemented in Go. It is fast and highly configurable.
	"github.com/microcosm-cc/bluemonday"
)

var HTMLSanitizer = bluemonday.UGCPolicy()

func init() {
	HTMLSanitizer.AllowStandardAttributes()
}

//	'HumanizeString' humanize separates string based on capitalize letters
//	e.g. "OrderItem" -> "Order Item"
func HumanizeString(str string) string {
	var human []rune
	for i, l := range str {
		if i > 0 && isUppercase(byte(l)) {
			if (!isUppercase(str[i-1]) && str[i-1] != ' ') || (i+1 < len(str) && !isUppercase(str[i+1])) && str[i+1] != ' ' && str[i-1] != ' ' {
				human = append(human, rune(' '))
			}
		}

		human = append(human, l)
	}

	return strings.Title(string(human))
}

func isUppercase(char byte) bool {
	return 'A' <= char && char <= 'Z'
}

//	'ToParamString' replaces spaces and sparaters words (by uppercase letters) with
//	underscores in a string, also downcase it
//	e.g. "ToParamString" -> "to_param_string", "To ParamString" -> "to_param_string"
func ToParamString(str string) string {
	return gorm.ToDBName(strings.Replace(str, " ", "_", -1))
}

//	'PatchURL' updates thr query part of the request url.
//	e.g. PatchURL("google.com","key","value") -> "google.com?key=value"
func PatchURL(originalURL string, params ...interface{}) (patchedURL string, err error) {
	urlResult, err := url.Parse(originalURL)
	if err != nil {
		return
	}

	query := urlResult.Query()
	for i := 0; i < len(params)/2; i++ {
		// Check if params is key&value pair
		key := fmt.Sprintf("%v", params[i*2])
		value := fmt.Sprintf("%v", params[i*2+1])
		if value == "" {
			query.Del(key)
		} else {
			query.Set(key, value)
		}
	}

	urlResult.RawQuery = query.Encode()
	patchedURL = urlResult.String()
	return
}

//	'JoinURL' updates the path part of the request url
//	e.g. JoinURL("google.com", "admin") => "google.com/admin"
//	e.g. JoinURL("google.com?q=keyword", "admin") => "google.com/admin?q=keyword"
func JoinURL(originalURL string, paths ...interface{}) (joinedURL string, err error) {
	urlReuslt, err := url.Parse(originalURL)
	if err != nil {
		return
	}

	urlPaths := []string{urlReuslt.Path}
	for _, p := range paths {
		urlPaths = append(urlPaths, fmt.Sprintf(p))
	}

	urlReuslt.Path = path.Join(urlPaths...)
	joinedURL = urlReuslt.String()
	return
}

// func SetCookie(cookie http.Cookie, context &)
