package utils

import (
	"database/sql/driver"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	//  The fantastic ORM library for Golang, aims to be developer friendly.
	"github.com/jinzhu/gorm"
	//  Now is a time toolkit for golang
	"github.com/jinzhu/now"
	//	a HTML sanitizer implemented in Go. It is fast and highly configurable.
	"github.com/microcosm-cc/bluemonday"

	"github.com/Sky-And-Hammer/TM_EC"
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
		urlPaths = append(urlPaths, fmt.Sprint(p))
	}

	urlReuslt.Path = path.Join(urlPaths...)
	joinedURL = urlReuslt.String()
	return
}

//	'SetCookie' set cookie for context
func SetCookie(cookie http.Cookie, context *TM_EC.Context) {
	cookie.HttpOnly = true
	if context.Request != nil && context.Request.URL.Scheme == "https" {
		cookie.Secure = true
	}

	if cookie.Path == "" {
		cookie.Path = "/"
	}

	http.SetCookie(context.Writer, &cookie)
}

//	'Stringify' stringify any data, if it is a struct, will try to use its Name, title, Code field, else will use its primary key
func Stringify(object interface{}) string {
	if obj, ok := object.(interface {
		stringify() string
	}); ok {
		return obj.stringify()
	}

	scope := gorm.Scope{Value: object}
	for _, column := range []string{"Name", "Title", "Code"} {
		if field, ok := scope.FieldByName(column); ok {
			result := field.Field.Interface()
			if valuer, ok := result.(driver.Valuer); ok {
				if result, err := valuer.Value(); err == nil {
					return fmt.Sprint(result)
				}
			}

			return fmt.Sprint(result)
		}
	}

	if scope.PrimaryField() != nil {
		if scope.PrimaryKeyZero() {
			return ""
		}
		return fmt.Sprintf("%v#%v", scope.GetModelStruct().ModelType.Name(), scope.PrimaryKeyValue())
	}

	return fmt.Sprint(reflect.Indirect(reflect.ValueOf(object)).Interface())
}

//	'ModelType' get value's model type
func ModelType(value interface{}) reflect.Type {
	reflectType := reflect.Indirect(reflect.ValueOf(value)).Type()
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}

	return reflectType
}

//	'ParseTagOption' parse tag options to hash
func ParseTagOption(str string) map[string]string {
	tags := strings.Split(str, ";")
	setting := map[string]string{}
	for _, value := range tags {
		v := strings.Split(value, ":")
		k := strings.TrimSpace(strings.ToUpper(v[0]))
		if len(v) == 2 {
			setting[k] = v[1]
		} else {
			setting[k] = k
		}
	}

	return setting
}

//	'ExistWithMsg' debug error messages and print stack
func ExistWithMsg(msg interface{}, value ...interface{}) {
	fmt.Printf("\n"+filenameWithLineNum()+"\n"+fmt.Sprint(msg)+"\n", value...)
	debug.PrintStack()
}

//	'FileServer' file server that disabled file listing
func FileServer(dir http.Dir) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Join(string(dir), r.URL.Path)
		if f, err := os.Stat(p); err == nil && !f.IsDir() {
			http.ServeFile(w, r, p)
			return
		}

		http.NotFound(w, r)
	})
}

func filenameWithLineNum() string {
	var total = 10
	var results []string
	for i := 2; i < 15; i++ {
		if _, file, line, ok := runtime.Caller(i); ok {
			total--
			results = append(results[:0], append([]string{fmt.Sprintf("%v:%v", strings.TrimPrefix(file, os.Getenv("GOPATH")+"src/"), line)}, results[0:]...)...)
			if total == 0 {
				return strings.Join(results, "\n")
			}
		}
	}

	return ""
}

//	'GetLocale' get locale from reqeust, cookie, after get the locale, will write the locale to the cookie if possible
//	Overwrite the default logic with
//		utils.GetLocal = func(context *TM_EC.Context) string {
//			//	.....
//		}
var GetLocale = func(context *TM_EC.Context) string {
	if locale := context.Request.Header.Get("Locale"); locale != "" {
		return locale
	}

	if locale := context.Request.URL.Query().Get("locale"); locale != "" {
		if context.Writer != nil {
			context.Request.Header.Set("Locale", locale)
			SetCookie(http.Cookie{Name: "locale", Value: locale, Expires: time.Now().AddDate(1, 0, 0)}, context)
		}

		return locale
	}

	if locale, err := context.Request.Cookie("locale"); err == nil {
		return locale.Value
	}

	return ""
}

//	'ParseTime' parse time from string
//	Overwrite the default logic with
//		utils.ParseTime = func(timeStr string, context *TM_EC.Context) (time.Time, error) {
//			//	.....
//		}
var ParseTime = func(timeStr string, context *TM_EC.Context) (time.Time, error) {
	return now.Parse(timeStr)
}

//	'FormatTime' format time to string
//	Overwrite the default logic with
//		utils.FormatTime = func(date time.Time, format string, context *TM_EC.Context) string {
//			//	.....
//		}
var FormatTime = func(date time.Time, format string, context *TM_EC.Context) string {
	return date.Format(format)
}
