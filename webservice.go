package main

import "fmt"
import "log"
import "os"
import "os/exec"
import "net/http"
import "strings"
import "io/ioutil"
//import "strconv"
import "encoding/base64"
//import "encoding/hex"
//import "crypto/md5"
import "time"
import "math/rand"
import "errors"
import "net/url"
import "database/sql"

import _ "github.com/go-sql-driver/mysql"

import "github.com/losfair/bdoauth-go/bdoauth"
import "github.com/losfair/bdtts-go/bdtts"
import "github.com/losfair/bdsr-go/bdsr"

import "tzacmcheck"
import "timeline"
import "scsapi"

var bdToken string

var logBuf string

var tzacmCheck_count uint = 0
var wxCollect_count uint = 0

func onTZACMCheckRequest(w http.ResponseWriter,r *http.Request) {
	if tzacmCheck_count>=3 {
		w.Write([]byte("Too many requests now. Please wait"))
		return
	}
	tzacmCheck_count++
	defer (func() {
		tzacmCheck_count--
	})()
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)!=4 {
		w.Write([]byte("Error: Bad request"))
		return
	}
	w.Write([]byte(tzacmcheck.DoCheck(parts[2],parts[3])))
}

func onWXCollectRequest(w http.ResponseWriter, r *http.Request) {
	if wxCollect_count>=3 {
		w.Write([]byte("Too many requests now. Please wait"))
		return
	}
	wxCollect_count++
	defer (func() {
		wxCollect_count--
	})()
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)!=3 {
		w.Write([]byte("Error: Bad request"))
		return
	}
//	salt := r.RemoteAddr+"###"+strconv.Itoa(int(time.Now().UnixNano()))+"###"+strconv.Itoa(rand.Intn(10000000))
/*	salt := parts[2]
	m := md5.New()
	m.Write([]byte(salt))
	resKey := hex.EncodeToString(m.Sum(nil))
*/
	resKey := parts[2]
	if len(resKey)!=8 {
		w.Write([]byte("Bad resKey: length"))
		return
	}
	for i:=0;i<len(resKey);i++ {
		if !(resKey[i]>='0' && resKey[i]<='9') && !(resKey[i]>='a' && resKey[i]<='z') {
			w.Write([]byte("Bad resKey: Invalid characters"))
			return
		}
	}
	cmd := exec.Command("python","python/accountInfoCollector.py",resKey)
//	cmd.Output()
	err := cmd.Start()
	if err!=nil {
		w.Write([]byte("exec failed: "+err.Error()))
		return
	}
	timer := time.AfterFunc(30*time.Second,func() {
		cmd.Process.Kill()
//		data,_ := ioutil.ReadAll(cmdOut)
		w.Write([]byte("Timeout. "))
		log.Println("exec timeout for resKey",resKey)
//		log.Println(data)
	})
	cmd.Wait()
	timer.Stop()
	w.Write([]byte(resKey))
}

func logFromPC(w http.ResponseWriter, r *http.Request) {
	fmt.Println("[*] logFromPC")
	if strings.Split(r.RemoteAddr,":")[0]!="172.16.9.14" {
		w.Write([]byte("Permission denied"))
		return
	}
	parts := strings.SplitN(r.URL.Path,"/",3)
	if len(parts)!=3 {
		w.Write([]byte("Bad request"))
		return
	}
	v,err := base64.StdEncoding.DecodeString(parts[2])
	if err!=nil {
		w.Write([]byte("Bad request"))
		return
	}
	logBuf+=string(v)+"\n"
}

func showLog(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(logBuf))
}

func onTTSRequest(w http.ResponseWriter, r *http.Request) {
	if bdToken=="" {
		w.Write([]byte("bdToken not initialized."))
		return
	}
	parts := strings.Split(r.URL.Path,"/")
	if len(parts) != 3 {
		w.Write([]byte("Bad request"))
		return
	}
	args,err := url.ParseQuery(parts[2])
	if err!=nil {
		w.Write([]byte("Bad request"))
		return
	}

	t,ok := args["text"]
	if !ok {
		w.Write([]byte("Bad request"))
		return
	}

	text := t[0]

	result,err := bdtts.Request(bdToken,text)
	if err!=nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type","audio/mp3")
	w.Write(result)
}

func onAudioChannel(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)<3 {
		w.Write([]byte("Bad request"))
		return
	}
	action := parts[2]

	switch action {
		case "sr":
			defer r.Body.Close()
			reqData,err := ioutil.ReadAll(r.Body)
			if err!=nil {
				w.Write([]byte("Unable to read request data: "+err.Error()))
				return
			}
			text,err := bdsr.Request(bdToken,reqData)
			if err!=nil {
				w.Write([]byte(err.Error()))
				return
			}
			w.Write([]byte(text))
		default:
			w.Write([]byte("Not supported"))
	}
}

func prepareSCS() {
	confData,err := ioutil.ReadFile("scscfg.txt")
	if err!=nil {
		log.Fatal("SCS initialization failed")
	}
	conf := strings.TrimSpace(string(confData))
	parts := strings.Split(conf,":")
	if len(parts)!=3 {
		log.Fatal("Bad SCS configuration")
	}
	scsapi.SetAccessKey(parts[0])
	scsapi.SetSecretKey(parts[1])
	scsapi.SetBucketName(parts[2])
}

func initBDOAuth(id,secret string) (string,error) {
	ret := bdoauth.RequestClientCredentials(id,secret)
	if ret==nil {
		return "",errors.New("bdoauth request failed")
	}
	tk,ok := ret["access_token"]
	if !ok {
		return "",errors.New("Unable to get access token")
	}
	return tk.(string),nil
}

func main() {
//	logBuf = ""

	os.Setenv("QT_QPA_PLATFORM","offscreen")

	rand.Seed(time.Now().UnixNano())

	sqlConnString,err := ioutil.ReadFile("sqlConnString.txt")
	if err!=nil {
		panic(err)
	}
	db,err := sql.Open("mysql",string(sqlConnString))
	if err!=nil {
		panic(err)
	}

	err = db.Ping()
	if err!=nil {
		panic(err)
	}

	db.SetMaxOpenConns(16)
	db.SetConnMaxLifetime(10*time.Second)

	prepareSCS()

	timeline.Prepare("timeline_static.json.gz",db)

	bdToken = ""

	var bdAPIKeyParts []string

	bdAPIKey,err := ioutil.ReadFile("bdAPIKey.txt")
	if err!=nil {
		log.Println("WARNING: No bdAPIKey.txt. Skipping Baidu OAuth initialization.")
		goto startListen
	}

	bdAPIKeyParts = strings.Split(string(bdAPIKey),":")
	if len(bdAPIKeyParts)!=2 {
		log.Println("Bad format for bdAPIKey.txt, please check. Skipping Baidu OAuth initialization.")
		goto startListen
	}

	bdToken,err = initBDOAuth(strings.TrimSpace(bdAPIKeyParts[0]),strings.TrimSpace(bdAPIKeyParts[1]))
	if err!=nil {
		log.Println("WARNING: Unable to init Baidu OAuth:",err.Error())
	}

	startListen:

	listenAddr := ":6084"
	fmt.Println("Listening on",listenAddr)

	http.HandleFunc("/tzACMCheck/",onTZACMCheckRequest)
	http.HandleFunc("/wxCollect/",onWXCollectRequest)
//	http.HandleFunc("/logFromPC/",logFromPC)
//	http.HandleFunc("/showLog/",showLog)
	http.HandleFunc("/ttsRequest/",onTTSRequest)
	http.HandleFunc("/audioChannel/",onAudioChannel)
	http.HandleFunc("/timeline/",timeline.OnRequest)

	http.ListenAndServe(listenAddr,nil)
}
