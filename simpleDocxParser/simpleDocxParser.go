package simpleDocxParser

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

type Document struct {
	originalPath string
	tempPath     string

	parsedDocument xmlDocument
	styles map[string]string
}

var htmlElementAliases = map[string]string {
		"title": "<h1>%s</h1>",
		"heading 1": "<h2>%s</h2>",
		"heading 2": "<h3>%s</h3>",
		"heading 3": "<h4>%s</h4>",
		"heading 4": "<h5>%s</h5>",
}

func New(file string) (doc Document, err error) {
	doc.styles = make(map[string]string)

	md5hasher := md5.New()
	md5hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	md5hasher.Write([]byte(file))

	doc.originalPath = file
	doc.tempPath = path.Join("/tmp/docx2cleanhtml/", hex.EncodeToString(md5hasher.Sum(nil)))

	folderErr := os.MkdirAll(doc.tempPath, 0750)
	zipReader, zipErr := zip.OpenReader(file)

	if zipErr == nil {
		if folderErr == nil {
			for _, file := range zipReader.File {
				fmt.Println(file.Name)
				if isAcceptedFile(file.Name) {
					ofHandle, ofErr := os.OpenFile(path.Join(doc.tempPath, path.Base(file.Name)), os.O_WRONLY|os.O_CREATE, 0750)
					fdHandle, fdErr := file.Open()
					if fdErr == nil {
						if ofErr == nil {
							_, copyErr := io.Copy(ofHandle, fdHandle)
							if copyErr != nil {
								log.Fatal(copyErr.Error())
							}
							ofcErr := ofHandle.Close()
							if ofcErr != nil {
								log.Fatal(ofcErr.Error())
							}
						} else {
							err = ofErr
						}
						fccErr := fdHandle.Close()
						if fccErr != nil {
							log.Fatal(fccErr.Error())
						}
					} else {
						log.Fatal(fdErr.Error())
					}
				}
			}
		} else {
			log.Fatal(folderErr.Error())
		}
	} else {
		log.Fatal(zipErr.Error())
	}

	return doc, err
}

func isAcceptedFile(filename string) bool {
	requiredFiles := []string{
		"word/document.xml",
		"word/styles.xml",
		"word/_rels/document.xml.rels",
	}
	for _, elem := range requiredFiles {
		if filename == elem {
			return true
		}
	}

	return false
}

func (doc *Document) ReadRelations() {
	doc.readDocuments()
	doc.readStyles()
	doc.close()
}

func (doc *Document) readStyles() {
	file, fileErr := os.Open(path.Join(doc.tempPath, "styles.xml"))
	var parsedStyles xmlStyles
	if fileErr == nil {
		readAllContent, readAllErr := ioutil.ReadAll(file)
		if readAllErr == nil {
			parseErr := xml.Unmarshal(readAllContent, &parsedStyles)
			if parseErr != nil {
				log.Fatal(parseErr)
			}

			for _, style := range parsedStyles.Xstyles {
				if style.XstyleId != "" {
					doc.styles[style.XstyleId] = style.XstyleId
				}
			}
		}
	}
}

func (doc *Document) readDocuments() {
	file, fileErr := os.Open(path.Join(doc.tempPath, "document.xml"))
	if fileErr == nil {
		byteContent, readAllErr := ioutil.ReadAll(file)
		if readAllErr == nil {
			parseErr := xml.Unmarshal(byteContent, &doc.parsedDocument)
			if parseErr != nil {
				log.Fatal(parseErr)
			}
		} else {
			log.Fatal(readAllErr)
		}
	} else {
		log.Fatal(fileErr.Error())
	}
}

func (doc *Document) readParagraphs(relativePath string) {

}

func (doc *Document) close() (err error) {
	return os.RemoveAll(doc.tempPath)
}

/*func (doc *Document) GetHTML() {

	htmlOut := ""

	for _, pg := range doc.parsedDocument.Xbody.Xparagraphs {
		for _, subPg := range pg.Xr {
			htmlOut = htmlOut + fmt.Sprintf("%v", subPg.Xt)
		}
	}
}*/