package turingbot

import "net/http"
import "net/url"
import "encoding/json"
import "io/ioutil"
import "strings"

func Request(userid string,msg string) string {
	v := url.Values{}
	v.Set("key","410855860241ab02024b7cf557f84f22")
	v.Set("info",msg)
	v.Set("userid",userid)
	res,err := http.Post("http://www.tuling123.com/openapi/api","application/x-www-form-urlencoded",strings.NewReader(v.Encode()))
	if err!=nil {
		return err.Error()
	}
	body,err := ioutil.ReadAll(res.Body)
	if err!=nil {
		return err.Error()
	}

	var ret map[string]interface{}

	err = json.Unmarshal([]byte(body),&ret)
	if err!=nil {
		return err.Error()
	}

	val,ok := ret["text"]
	if !ok {
		return "No text received."
	}

	switch val.(type) {
		case string:
			return val.(string)
		default:
			return "Unknown response type."
	}
}
