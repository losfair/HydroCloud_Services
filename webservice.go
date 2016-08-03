package main

import "fmt"
import "log"
import "os"
import "os/exec"
import "net/http"
import "strings"
//import "io/ioutil"
//import "strconv"
import "encoding/base64"
//import "encoding/hex"
//import "crypto/md5"
import "time"
//import "math/rand"

import "tzacmcheck"

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

func main() {
	logBuf = ""

	os.Setenv("QT_QPA_PLATFORM","offscreen")

	listenAddr := ":6084"
	fmt.Println("Listening on",listenAddr)

	http.HandleFunc("/tzACMCheck/",onTZACMCheckRequest)
	http.HandleFunc("/wxCollect/",onWXCollectRequest)
	http.HandleFunc("/logFromPC/",logFromPC)
	http.HandleFunc("/showLog/",showLog)

	http.ListenAndServe(listenAddr,nil)
}
