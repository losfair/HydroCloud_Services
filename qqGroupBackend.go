package main

import "log"
import "time"
import "strings"
import "strconv"
import "net/http"
import "io/ioutil"
import "database/sql"
import "encoding/json"
import "math/rand"

//import "zhuji"
import _ "github.com/go-sql-driver/mysql"

// #cgo LDFLAGS: -L. -lLetsChat
// #include "c/letsChat/detect.h"
import "C"

var sqlConn *sql.DB

var stmtQueryGroupSign *sql.Stmt
var stmtInsertGroupSign *sql.Stmt

var zhujiMsgChannels map[int64]chan string
var zhujiOutputChannels map[int64]chan string

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

func initGlobalMaps() {
	zhujiMsgChannels = make(map[int64]chan string)
	zhujiOutputChannels = make(map[int64]chan string)
}

func prepareStmt() {
	var err error

	stmtQueryGroupSign,err = sqlConn.Prepare("SELECT id FROM group_sign WHERE gid = ? AND uid = ? AND sign_time >= ?")
	checkError(err)

	stmtInsertGroupSign,err = sqlConn.Prepare("INSERT INTO group_sign (gid,uid,sign_time) VALUES(?,?,?)")
	checkError(err)
}

/*
func handleZhuji(gid int64, req string) string {
	needStart := false
	msgChan,ok := zhujiMsgChannels[gid]
	if !ok {
		zhujiMsgChannels[gid]=make(chan string)
		msgChan = zhujiMsgChannels[gid]
		needStart=true
	}
	outChan,ok := zhujiOutputChannels[gid]
	if !ok {
		zhujiOutputChannels[gid] = make(chan string)
		outChan = zhujiOutputChannels[gid]
		needStart=true
	}
	log.Println("New Zhuji request from group",gid,"-",req)
	if needStart {
		log.Println("Starting goroutine for group",gid)
		go zhuji.Start(msgChan,outChan)
	}
	msgChan <- req+"\n"
	return <-outChan
}
*/

func doGroupSign(ginfo,uinfo map[string]interface{}) string {
	timeData := time.Now()

	dayStart,_ := time.Parse("2006/01/02",timeData.Format("2006/01/02"))
	timeStampStart := dayStart.Unix()

	groupAccount := strconv.FormatInt(int64(ginfo["account"].(float64)),10)
	userName := uinfo["nick"].(string)

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

	log.Println("GID",int64(requestData["ginfo"].(map[string]interface{})["account"].(float64)),"-",userMessage)

	apiReturnString := ""

/*	if len(userMessage)>2 && userMessage[0:1]=="." {
		apiReturnString = handleZhuji(int64(requestData["ginfo"].(map[string]interface{})["account"].(float64)),userMessage[1:len(userMessage)])
		goto writeOut
	}
*/

	groupId := int64(requestData["ginfo"].(map[string]interface{})["account"].(float64))

	C.chatMsgInput(C.CString(userMessage),C.id_type(groupId))

/*	if C.chatGetTotalWeight(C.id_type(groupId))>=1.0 {
		C.chatClearTotalWeight(C.id_type(groupId))
		w.Write([]byte("这群风气不太对啊。"))
		return
	}
*/

	letsChatReturn := C.GoString(C.chatGetOutputText(C.id_type(groupId)))
	if letsChatReturn != "OK" {
		w.Write([]byte(letsChatReturn))
		return
	}

	switch userMessage {
		case "签到":
			apiReturnString = doGroupSign(requestData["ginfo"].(map[string]interface{}),requestData["uinfo"].(map[string]interface{}))
		case "抛骰子":
			apiReturnString = generateTouZi()
		case "+1s":
			apiReturnString = "已禁用。"
		case "续一秒":
			apiReturnString = "已禁用。"
		default:
			apiReturnString = "UNSUPPORTED"
	}

//	writeOut:

	w.Write([]byte(apiReturnString))
}

func main() {
	sqlConn = initDatabase()

	initGlobalMaps()

	prepareStmt()

	http.HandleFunc("/newMessage/",onNewMessage)
	http.ListenAndServe(":6086",nil)
}
