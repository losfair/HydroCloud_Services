package timeline

import "os"
import "log"
import "fmt"
import "time"
import "strconv"
import "strings"
import "math/rand"
import "io/ioutil"
import "net/http"
import "encoding/json"
import "encoding/base64"
import "compress/gzip"
import "database/sql"
import "crypto/md5"
import "errors"
//import "image"
import "image/png"
import "image/jpeg"
import "bytes"

import "github.com/losfair/scsapi-go/scsapi"

var urlPrefix string

var staticCache map[string][]byte

var tplGetUpdates *sql.Stmt
var tplCreateUpdate *sql.Stmt
var tplPlusoneGet *sql.Stmt
var tplPlusoneUpdate *sql.Stmt

func getUpdates(lastId int64) []map[string]interface{} {
	allUpdates := make([]map[string]interface{},0)

	res,err := tplGetUpdates.Query(lastId)

	if err!=nil {
		return allUpdates
	}

	defer res.Close()

	for res.Next() {
		currentId := 0
		imgURL := ""
		comments := ""
		plusone := 0
		updateTime := 0
		err = res.Scan(&currentId,&imgURL,&comments,&plusone,&updateTime)
		if err!=nil {
			return allUpdates
		}
		update := make(map[string]interface{})
		update["id"] = currentId
		update["imgURL"] = imgURL
		update["comments"] = comments
		update["plusone"] = plusone
		update["time"] = updateTime
		allUpdates = append(allUpdates,update)
	}

	return allUpdates
}

func checkImageURL(targetURL string) bool {
	if len(targetURL) < 10 || len(targetURL) >= 256 {
		return false
	}
	if !strings.HasPrefix(targetURL,"http://") && !strings.HasPrefix(targetURL,"https://") {
		return false
	}
	return true
}

func createUpdate(w http.ResponseWriter, r *http.Request) {
	fileReader,fileHeader,err := r.FormFile("img")
	if err!=nil {
		w.WriteHeader(400)
		w.Write([]byte("Upload required"))
		return
	}
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType=="" {
		w.WriteHeader(400)
		w.Write([]byte("Content-Type header required"))
		return
	}
	fileData,err := ioutil.ReadAll(fileReader)
	if err!=nil {
		w.WriteHeader(400)
		w.Write([]byte("Unable to read the upload file"))
		return
	}
	dataReader := bytes.NewReader(fileData)
	fileSuffix := ".bin"
	if contentType == "image/png" {
		_,err = png.Decode(dataReader)
		fileSuffix = ".png"
	} else if(contentType == "image/jpeg") {
		_,err = jpeg.Decode(dataReader)
		fileSuffix = ".jpg"
	} else {
		err = errors.New("Bad content type")
	}
	if err!=nil {
		w.WriteHeader(500)
		w.Write([]byte("Unable to decode image: "+err.Error()))
		return
	}
	md5Context := md5.New()
	md5Context.Write(fileData)
	md5String := fmt.Sprintf("%32x",md5Context.Sum(nil))
	fullPath := "/timeline_images/"+md5String+fileSuffix
	scsapi.Upload(fullPath,"","",fileData)
	targetURL := "https://hydrocloud.sinacloud.net"+fullPath
	_,err = tplCreateUpdate.Exec(string(targetURL),int(time.Now().Unix()))
	if err!=nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
//	lastInsertId,_ := result.LastInsertId()
	w.Write([]byte(targetURL))
}

func plusoneGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data,err := ioutil.ReadAll(r.Body)
	if err!=nil || len(data)>10 {
		w.WriteHeader(500)
		w.Write([]byte("Unable to read request body or invalid request"))
		return
	}
	targetId,err := strconv.ParseInt(string(data),10,32)
	if err!=nil {
		w.WriteHeader(400)
		w.Write([]byte("Bad request format"))
		return
	}
	res,err := tplPlusoneGet.Query(targetId)
	if err!=nil {
		w.WriteHeader(500)
		w.Write([]byte("Database query failed"))
		return
	}
	defer res.Close()

	ok := res.Next()
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte("Target id not found"))
		return
	}
	var plusoneCount int
	res.Scan(&plusoneCount)

	w.Write([]byte(strconv.Itoa(plusoneCount)))
}

func plusoneUpdate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data,err := ioutil.ReadAll(r.Body)
	if err!=nil || len(data)>10 {
		w.WriteHeader(500)
		w.Write([]byte("Unable to read request body or invalid request"))
		return
	}
	targetId,err := strconv.ParseInt(string(data),10,32)
	if err!=nil {
		w.WriteHeader(400)
		w.Write([]byte("Bad request format"))
		return
	}
	res,err := tplPlusoneGet.Query(targetId)
	if err!=nil {
		w.WriteHeader(500)
		w.Write([]byte("Database query failed"))
		return
	}
	defer res.Close()

	ok := res.Next()
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte("Target id not found"))
		return
	}
	var plusoneCount int
	res.Scan(&plusoneCount)

	tplPlusoneUpdate.Exec(plusoneCount+1,targetId)

	w.Write([]byte(strconv.Itoa(plusoneCount+1)))
}

var imgTokens map[string]string

func requestImageToken(w http.ResponseWriter, r *http.Request) {
	if imgTokens == nil {
		imgTokens = make(map[string]string)
	}

	referer := r.Header.Get("Referer")
	if !strings.HasPrefix(referer,urlPrefix) {
		w.WriteHeader(403)
		w.Write([]byte("Referer not allowed"))
		return
	}

	defer r.Body.Close()

	reqData,err := ioutil.ReadAll(r.Body)
	if err!=nil {
		return
	}

	if len(reqData)>128 {
		w.WriteHeader(400)
		w.Write([]byte("Request too large"))
		return
	}

	reqStr := string(reqData)

	parts := strings.Split(r.URL.Path,"/")
	if len(parts)!=5 {
		return
	}

	cmd := parts[4]

	switch cmd {
		case "get":
			if len(reqData)>9 {
				w.WriteHeader(400)
				w.Write([]byte("Bad key format"))
				return
			}

			data,ok := imgTokens[reqStr]
			if !ok {
				w.WriteHeader(404)
				w.Write([]byte("Item not found"))
				return
			}
			w.Write([]byte(data))
		case "set":
			keyData := make([]byte,8)
			for i:=0;i<8;i++ {
				keyData[i] = byte('a' + rand.Intn(100000)%26)
			}
			key := string(keyData)
			imgTokens[key] = reqStr
			w.Write(keyData)
		default:
			w.WriteHeader(400)
			w.Write([]byte("Bad request type"))
			return
	}

}

func onAJAXRequest(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)<4 {
		w.WriteHeader(500)
		w.Write([]byte("Bad AJAX request"))
		return
	}

	switch parts[3] {
		case "clientLog":
			if len(parts)!=5 {
				w.WriteHeader(400);
				w.Write([]byte("Bad request"))
				return
			}
			dataToLog,err := strconv.ParseInt(parts[4],10,32)
			if err!=nil {
				w.WriteHeader(400)
				w.Write([]byte("Bad dataToLog"))
				return
			}
			log.Printf("Client log [%s]: %d\n",r.RemoteAddr,dataToLog)
		case "getUpdates":
			if len(parts)!=5 {
				w.WriteHeader(400);
				w.Write([]byte("Bad request"))
				return
			}
			lastId,err := strconv.ParseInt(parts[4],10,32)
			if err!=nil {
				w.WriteHeader(400)
				w.Write([]byte("Bad lastId"))
				return
			}
			jsonData,err := json.Marshal(getUpdates(lastId))
			if err!=nil {
				w.WriteHeader(500)
				w.Write([]byte("Unable to get updates"))
				return
			}
			w.Write(jsonData)
		case "createUpdate":
			if len(parts)!=4 {
				w.WriteHeader(400);
				w.Write([]byte("Bad request"))
				return
			}
			createUpdate(w,r)
		case "requestImageToken":
			if len(parts)!=5 {
				w.WriteHeader(400);
				w.Write([]byte("Bad request"))
				return
			}
			requestImageToken(w,r)
		case "plusoneGet":
			if len(parts)!=4 {
				w.WriteHeader(400);
				w.Write([]byte("Bad request"))
				return
			}
			plusoneGet(w,r)
		case "plusoneUpdate":
			if len(parts)!=4 {
				w.WriteHeader(400);
				w.Write([]byte("Bad request"))
				return
			}
			plusoneUpdate(w,r)
		default:
			w.WriteHeader(404)
			w.Write([]byte("Method not found"))
	}
}

func onStaticRequest(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(r.URL.Path,"/",4)
	if len(parts)!=4 {
		w.WriteHeader(400)
		w.Write([]byte("Bad request"))
		return
	}

	data,ok := staticCache["/"+parts[3]]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte("File not found."))
		return
	}

	contentTypes := map[string]string {
		"js": "application/x-javascript",
		"css": "text/css",
		"html": "text/html",
		"svg": "image/svg+xml",
	}

	dotParts := strings.Split(parts[3],".")
	fileType := dotParts[len(dotParts)-1]

	contentType,ok := contentTypes[fileType]
	if !ok {
		w.WriteHeader(403);
		w.Write([]byte("Unknown file type."))
		return
	}

	w.Header().Set("Content-Type",contentType+";charset=utf-8")
	w.Header().Set("Cache-Control","max-age=120")

	w.Write(data)
}

func OnRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s\n",r.RemoteAddr,r.URL.Path)

	w.Header().Set("Server","HWS Timeline Module")
	parts := strings.Split(r.URL.Path,"/")
	if len(parts)<3 {
		w.WriteHeader(400)
		w.Write([]byte("Bad request"))
		return
	}
	switch parts[2] {
		case "static":
			onStaticRequest(w,r)
		case "ajax":
			onAJAXRequest(w,r)
		default:
			w.WriteHeader(404)
			w.Write([]byte("Request type not supported"))
	}
}

func Prepare(static_path string, db *sql.DB) {
	staticCache = make(map[string][]byte)
	tempCache := make(map[string]string)

	fileHandle,err := os.Open(static_path)
	if err!=nil {
		log.Fatal("[A]",err)
	}

	reader,err := gzip.NewReader(fileHandle)
	if err!=nil {
		log.Fatal("[B]",err)
	}
	jsonData,err := ioutil.ReadAll(reader)
	if err!=nil {
		log.Fatal("[C]",err)
	}

	err = json.Unmarshal(jsonData,&tempCache)

	if err!=nil {
		log.Fatal(err)
	}

	fileCount := 0

	for k,v := range tempCache {
		staticCache[k],err = base64.StdEncoding.DecodeString(v)
		if err!=nil {
			log.Fatal(err)
		}
		fileCount++
	}
	log.Printf("Static cache loaded, %d items.\n",fileCount)

	tplGetUpdates,_ = db.Prepare("SELECT id,imgURL,comments,plusone,time FROM timeline WHERE id > ?")
	tplCreateUpdate,_ = db.Prepare("INSERT INTO timeline (imgURL,comments,time) VALUES(?,\"\",?)")
	tplPlusoneGet,_ = db.Prepare("SELECT plusone FROM timeline WHERE id = ?")
	tplPlusoneUpdate,_ = db.Prepare("UPDATE timeline SET plusone = ? WHERE id = ?")
}
