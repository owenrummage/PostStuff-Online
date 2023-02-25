package main

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var maxUploadSize int64 = 2 * 1024 * 1024 // 2 mb
const uploadPath = "/data"

const domain = "https://postshit.online"
const port = "8080"

func main() {

	http.HandleFunc("/", fileHandler())

	log.Print("Server started on "+domain+", post to / for uploading files and /{fileName} for downloading")
	log.Fatal(http.ListenAndServe(":"+port, nil))
}


func fileHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if r.URL.Path == "/" {
				w.Write([]byte(`
WHAT?
------
This is a free file upload service hosted by Owen Rummage (rummage.cc)
This service can be used by issuing the following command in your terminal 

curl -F'file=@myfile.ext' https://postshit.online/

Ofc replace 'myfile.ext' with the name of your file or absolute path on your local computer.



INFORMATION
-----------
This service is posted free of charge for anyones use to store a quick file for others to access 
or for transferring files from one machine to another. The upload limit is set to 2mb for all users.
Have fun and obey the TOS.

				
TERMS OF SERVICE
----------------

You can NOT use this site for:
    * piracy
    * pornography and gore
    * extremist material of any kind
    * malware / botnet C&C
    * anything related to crypto currencies
    * backups
    * CI build artifacts
    * doxxing, database dumps containing personal information
    * anything illegal under US (Tennessee) law

Uploads found to be in violation of these rules will be removed,
and the originating IP address blocked from further uploads.

EXAMPLES
--------
 - https://postshit.online/PsychzNetworks.mp3
 - https://postshit.online/Quirked_Up_White_Boy.mp4
 - https://postshit.online/demo_page.html (yes you can even upload renderable html!)


Last Updated ${DATE}
`))
			}else{	
				fileBytes,err := ioutil.ReadFile(uploadPath + "/" + r.URL.Path)
				if(err != nil){
					renderError(w, "NO_SUCH_FILE", http.StatusNotFound)
				}
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Write(fileBytes)
			}
			return
		}

		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Printf("Could not parse multipart form: %v\n", err)
			renderError(w, "CANT_PARSE_FORM", http.StatusInternalServerError)
			return
		}

		// parse and validate file and post parameters
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}
		defer file.Close()
		// Get and print out file size
		fileSize := fileHeader.Size
		fmt.Printf("File size (bytes): %v\n", fileSize)

		if(r.Header.Get("Authorization") == "Joshua"){
			maxUploadSize = 1000 * 1024 * 1024	
		}

		// validate file size
		if fileSize > maxUploadSize {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}
		fileBytes, err := ioutil.ReadAll(file)
		if err != nil {
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		// Regex for file extension
		reg, err := regexp.Compile("\\.(.*)")
		if(err != nil){
			panic(err)
		}
		extension := reg.FindStringSubmatch(fileHeader.Filename)

		
		fileName := ""

		if(len(extension) > 0){
			fileName = randToken(12) + "."+extension[1]
		}else{
			fileName = randToken(12) + "_" + fileHeader.Filename	
		}

		if err != nil {
			renderError(w, "CANT_READ_FILE_TYPE", http.StatusInternalServerError)
			return
		}
		newPath := filepath.Join(uploadPath, fileName)
	
		if(r.Header.Get("Authorization") == "Joshua"){
			logger(time.Now().Format("2006.01.02 15:04:05")+"  "+"Joshua uploaded file! "+fileHeader.Filename+" : "+fileName)	
		}else{
			logger(time.Now().Format("2006.01.02 15:04:05")+"  "+"User uploaded file! "+fileHeader.Filename+" : "+fileName)
		}
		// write file
		newFile, err := os.Create(newPath)
		if err != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		defer newFile.Close() // idempotent, okay to call twice
		if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
			renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(domain + "/" +fileName))
	})
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func randToken(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func logger(line string){
	f, err := os.OpenFile("/var/log/postit.log", os.O_APPEND|os.O_WRONLY, 0644)
    	if err != nil {
        	fmt.Println(err)
        	return
    	}
    	_, err = fmt.Fprintln(f, line)
    	if err != nil {
        	fmt.Println(err)
                f.Close()
        	return
    	}
    	err = f.Close()
    	if err != nil {
        	fmt.Println(err)
        	return
    	}
}
