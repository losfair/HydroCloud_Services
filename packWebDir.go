package main

import "os"
import "fmt"
import "log"
import "strings"
import "io/ioutil"
import "path/filepath"
import "compress/gzip"
import "encoding/json"
import "encoding/base64"

func main() {
	if len(os.Args) != 3 {
		log.Fatal(fmt.Sprintf("Usage: %s [webdir] [outfile]",os.Args[0]))
	}

	static_path := os.Args[1]
	outFileName := os.Args[2]

	staticCache := make(map[string]string)

	fileCount := 0

	filepath.Walk(static_path, func(path string, fileInfo os.FileInfo, err error) error {
		if err!=nil {
			log.Fatal(err)
		}
		if fileInfo.IsDir() {
			return nil
		}

		data,err := ioutil.ReadFile(path)
		if err!=nil {
			log.Fatal(err)
		}

		parts := strings.SplitN(path,"/",2)
		if len(parts)!=2 {
			log.Fatal("Bad path",path)
		}

		staticCache["/"+parts[1]]=base64.StdEncoding.EncodeToString(data)
		fileCount++

		return nil
	})
	log.Printf("Static cache loaded, %d files.\n",fileCount)

	jsonData,err := json.Marshal(staticCache)

	if err!=nil {
		log.Fatal(err)
	}

	os.Create(outFileName)

	outFile,err := os.OpenFile(outFileName,os.O_WRONLY,0644)
	if err!=nil {
		log.Fatal(err)
	}

	writer := gzip.NewWriter(outFile)
	defer writer.Close()

	writer.Write(jsonData)
	writer.Flush()

	log.Println("Done.")
}
