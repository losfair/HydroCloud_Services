package main

import (
	"log"
	"time"
	"strings"
	"strconv"
	"net/http"
	"io/ioutil"
	"database/sql"
	"encoding/json"
	"math/rand"

	_ "github.com/go-sql-driver/mysql"

	"urlencoder"
	"chatserviceapi"
)

// #cgo LDFLAGS: -L. -lLetsChat -lstdc++
// #include "cpp/letsChatV2/api.h"
import "C"

var sqlConn *sql.DB

var stmtQueryGroupSign *sql.Stmt
var stmtInsertGroupSign *sql.Stmt

var groupBadWordLevels map[int64]int

func checkError(err error) {
	if err!=nil {
		log.Fatal(err)
	}
}

func initDatabase() *sql.DB {
	cfg,err := ioutil.ReadFile("sqlConnString_qqbot.txt")
	checkError(err)

	db,err := sql.Open("mysql",string(cfg))
	checkError(err)

	err = db.Ping()
	checkError(err)

	return db
}

func prepareStmt() {
	var err error

	stmtQueryGroupSign,err = sqlConn.Prepare("SELECT id FROM group_sign WHERE gid = ? AND uid = ? AND sign_time >= ?")
	checkError(err)

	stmtInsertGroupSign,err = sqlConn.Prepare("INSERT INTO group_sign (gid,uid,sign_time) VALUES(?,?,?)")
	checkError(err)
}

func doGroupSign(gid,uid int64) string {
	timeData := time.Now()

	dayStart,_ := time.Parse("2006/01/02",timeData.Format("2006/01/02"))
	timeStampStart := dayStart.Unix()

	groupAccount := strconv.FormatInt(gid,10)
	userName := strconv.FormatInt(uid,10)

	res := stmtQueryGroupSign.QueryRow(groupAccount,userName,timeStampStart)
	if res==nil {
		return "doGroupSign failed: QueryRow"
	}

	signDataID := 0

	err := res.Scan(&signDataID)

	if err==nil {
		return "你今天已经签到过了。"
	}

	execRes,err := stmtInsertGroupSign.Exec(groupAccount,userName,timeData.Unix())
	if err!=nil {
		return err.Error()
	}
	lastInsert,_ := execRes.LastInsertId()
	return "签到成功，签到数据行 ID: "+strconv.FormatInt(lastInsert,10)
}

func generateTouZi() string {
	return strconv.Itoa(rand.Intn(5)+1)
}

func tryChatservice(msg string) string {
	if strings.HasPrefix(msg,"安利") {
		parts := strings.SplitN(msg," ",2)
		if len(parts)!=2 {
			return chatserviceapi.Request("/getAllGroupNames")
		}
		return chatserviceapi.Request("/getGroupIntro/name="+urlencoder.EncodeComponent(parts[1]));
	}
	return ""
}

func getRequestInt(rd map[string]interface{},key string) int64 {
	rawData,ok := rd[key]
	if !ok {
		return 0
	}

	switch rawData.(type) {
		case int64:
			return rawData.(int64)
		case float64:
			return int64(rawData.(float64))
		default:
			return 0
	}


	return 0
}

func onNewMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	requestBody,err := ioutil.ReadAll(r.Body)
	if err!=nil {
		w.Write([]byte(err.Error()))
		return
	}

	requestData := make(map[string]interface{})
	err = json.Unmarshal(requestBody,&requestData)
	if err!=nil {
		w.Write([]byte(err.Error()))
		return
	}
	userMessageInterface,ok := requestData["content"]
	if !ok {
		w.Write([]byte("Bad value: content"))
		return
	}

	userMessage := ""

	switch userMessageInterface.(type) {
		case string:
			userMessage=userMessageInterface.(string)
		default:
			w.Write([]byte("Bad data type: content"))
			return
	}

	userMessage = strings.TrimSpace(userMessage)

	groupId := getRequestInt(requestData,"groupId")
	userId := getRequestInt(requestData,"userId")

	if groupId <= 0 || userId <= 0 {
		w.Write([]byte("Bad data type or value"))
		return
	}

	log.Printf("%d/%d - %s\n",groupId,userId,userMessage)

	apiReturnString := ""

	C.chatMsgInput(C.id_type(groupId),C.id_type(userId),C.CString(userMessage))

	badWordWeight := C.chatGetBadWordWeight(C.id_type(groupId))
	badWordLevel,ok := groupBadWordLevels[groupId]
	if !ok {
		badWordLevel = 0
	}

	if badWordWeight > 2.0  && badWordLevel < 2 {
		w.Write([]byte("这群药丸。"))
		groupBadWordLevels[groupId] = 2
		return
	} else if badWordWeight > 1.0 && badWordLevel < 1 {
		w.Write([]byte("这群风气不太对啊。"))
		groupBadWordLevels[groupId] = 1
		return
	} else if badWordLevel != 0 {
		groupBadWordLevels[groupId] = 0
	}

	switch userMessage {
		case "签到":
			apiReturnString = doGroupSign(groupId,userId)
		case "抛骰子":
			apiReturnString = generateTouZi()
		case "+1s":
			apiReturnString = "已禁用。"
		case "续一秒":
			apiReturnString = "已禁用。"
		default:
			apiReturnString = tryChatservice(userMessage)
			if apiReturnString == "" {
				apiReturnString = "UNSUPPORTED"
			}
	}

	w.Write([]byte(apiReturnString))
}

func main() {
	sqlConn = initDatabase()

	prepareStmt()

	groupBadWordLevels = make(map[int64]int)

	http.HandleFunc("/newMessage/",onNewMessage)
	http.ListenAndServe(":6086",nil)
}
