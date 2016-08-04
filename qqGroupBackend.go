package main

import "log"
import "time"
import "strconv"
import "net/http"
import "io/ioutil"
import "database/sql"
import "encoding/json"
import _ "github.com/go-sql-driver/mysql"

var sqlConn *sql.DB

var stmtQueryGroupSign *sql.Stmt
var stmtInsertGroupSign *sql.Stmt


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

	apiReturnString := ""

	switch userMessage {
		case "签到":
			apiReturnString = doGroupSign(requestData["ginfo"].(map[string]interface{}),requestData["uinfo"].(map[string]interface{}))
		default:
			apiReturnString = "Unsupported userMessage"
	}
	w.Write([]byte(apiReturnString))
}

func main() {
	sqlConn = initDatabase()

	prepareStmt()

	http.HandleFunc("/newMessage/",onNewMessage)
	http.ListenAndServe(":6086",nil)
}
