package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Request schema for handling POST method
type LinkShortenerRequest struct {
	Link	string
}

// Link Shortener service
type LinkShortener interface {
	lookup(slink string) string
	shorten(link string) string
}

// Hash Cut method
type HashCutLinkShortener struct {
	Length	int
	storage	LinkShortenerStorage
	salt	string
}

// Helper: Generate the hash
func (hc *HashCutLinkShortener) _hashCut(link string) string {
	h := sha256.New()
	h.Write([]byte(link))
	hash := h.Sum(nil)
	return fmt.Sprintf("%x", hash)[:hc.Length]
}

// Shortener utility function
func (hc *HashCutLinkShortener) shorten(link string) string {
	modifiedLink := link
	slink := hc._hashCut(modifiedLink);
	// Retry if it's duplicated
	for hc.storage.get(slink) != "" {
		modifiedLink = hc.salt + modifiedLink
		slink = hc._hashCut(modifiedLink)
	}
	hc.storage.add(slink, link)
	return slink
}

// Helper: Restores the link by removing salt
func (hc *HashCutLinkShortener) _restoreLink(link string) string {
	for strings.Contains(link, hc.salt) {
		link = link[len(hc.salt):]
	}
	return link
}

// Restores and returns the associated link to the slink
func (hc *HashCutLinkShortener) lookup(slink string) string {
	return hc._restoreLink(hc.storage.get(slink))
}

// Storage interface
type LinkShortenerStorage interface {
	get(slink string) string
	add(slink, link string)
}

// Simple Storage Provider that uses variables (RAM)
type SimpleStorageProvider struct {
	db	map[string]string
}

// Returns the link associated to the slink
func (ssp *SimpleStorageProvider) get(slink string) string {
	if link, ok := ssp.db[slink]; ok {
		return link
	}
	return ""
}

// Set the key/value pair: link -> slink
func (ssp *SimpleStorageProvider) add(slink, link string) {
	ssp.db[slink] = link
}

// Mysql Storage Provider that uses database
type MysqlStorageProvider struct {
	dsn	string
	db 	*gorm.DB
}

type GormShortenedLink struct {
	gorm.Model
	Slink	string	`gorm:"primaryKey"`
	Link	string
}

func NewMysqlStorageProvider(dsn string) *MysqlStorageProvider {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	db.AutoMigrate(&GormShortenedLink{})

	return &MysqlStorageProvider{dsn, db}
}

func (msp *MysqlStorageProvider) add(slink, link string) {
	msp.db.Create(&GormShortenedLink{
		Slink: slink,
		Link: link,
	})
}

func (msp *MysqlStorageProvider) get(slink string) string {
	var shortenedLink GormShortenedLink
	err := msp.db.First(&shortenedLink, "slink = ?", slink).Error

	if err == gorm.ErrRecordNotFound {
		println("error")
		return ""
	} else {
		return shortenedLink.Link
	}
}

func main () {
	r := mux.NewRouter()
	godotenv.Load()
	
	// storage := &SimpleStorageProvider{
	// 	db: make(map[string]string),
	// }

	storage := NewMysqlStorageProvider(os.Getenv("DSN"))

	ls := &HashCutLinkShortener{
		salt: "!",
		Length: 6,
		storage: storage,
	}

	r.HandleFunc("/lnk/{slink}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slink := vars["slink"]

		if link := ls.lookup(slink); link != "" {
			fmt.Fprintf(w, "%s", link)
			return
		}
		
		http.Error(w, "Invalid shortened link!", http.StatusBadRequest)
	}).Methods("GET")

	r.HandleFunc("/gen", func(w http.ResponseWriter, r *http.Request) {
		var data LinkShortenerRequest
		
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		link := data.Link
		slink := ls.shorten(link)
		
		fmt.Fprintf(w, "http://localhost:3000/lnk/%s\n", slink)

	}).Methods("POST")

	http.ListenAndServe(":3000", r)
}
