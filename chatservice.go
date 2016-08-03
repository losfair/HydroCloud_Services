package main

import "fmt"
import "net/http"
import "net/url"
import "database/sql"
import "strings"
import "io/ioutil"

import _ "github.com/go-sql-driver/mysql"

import "turingbot"

var tplGetGroupIntroByName *sql.Stmt

func getGroupIntro(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)!=3 {
		w.Write([]byte("Bad request"))
		return
	}
	args,err := url.ParseQuery(parts[2])
	if err!=nil {
		w.Write([]byte(err.Error()))
		return
	}
	name,ok := args["name"]
	if !ok {
		w.Write([]byte("Bad arguments or function not implemented."))
		return
	}
	if name[0]=="usage" {
		w.Write([]byte("用法：[HCBot] 安利 [社团名]"))
		return
	}
	res,err := tplGetGroupIntroByName.Query("%"+name[0]+"%")
	if err!=nil {
		w.Write([]byte(err.Error()))
		return
	}
	var content string
	res.Next()
	res.Scan(&content)
	if content=="" {
		w.Write([]byte("未找到此社团，请先在 https://apps.ixservices.net/anli/ 提交。"))
		return
	}
	w.Write([]byte(content))
}

func turingBot(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)!=3 {
		w.Write([]byte("Bad request #1"))
		return
	}
	args,err := url.ParseQuery(parts[2])
	if err!=nil {
		w.Write([]byte(err.Error()))
		return
	}
	userid,ok := args["userid"]
	if !ok {
		w.Write([]byte("Bad request #2: userid"))
		return
	}
	msg,ok := args["msg"]
	if !ok {
		w.Write([]byte("Bad request #3: msg"))
		return
	}
	w.Write([]byte(turingbot.Request(userid[0],msg[0])))
}

func main() {
	listenAddr := "127.0.0.1:6083"
	sqlConnString,err := ioutil.ReadFile("sqlConnString.txt")
	if err!=nil {
		panic(err)
	}
	db,err := sql.Open("mysql",string(sqlConnString))
	if err!=nil {
		panic(err)
	}

	tplGetGroupIntroByName,err = db.Prepare("SELECT content FROM intro_to_ntzx_groups_2016 WHERE name LIKE ? AND is_showed=1 ORDER BY id DESC")
	if err!=nil {
		panic(err)
	}

	http.HandleFunc("/getGroupIntro/",getGroupIntro)
	http.HandleFunc("/turingBot/",turingBot)

	fmt.Println("Listening on",listenAddr)
	http.ListenAndServe(listenAddr,nil)
}
