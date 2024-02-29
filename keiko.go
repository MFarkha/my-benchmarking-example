package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/chai2010/webp"
	_ "github.com/mattn/go-sqlite3"

	// _ "github.com/mxk/go-sqlite/sqlite3"

	"keiko/keikodb"
)

type WebpImage = bytes.Buffer

// Scans the given directory and returns all entries.
func getImageNames(path string) []os.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	return entries
}

// Read raw image data from disk.
func loadImg(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

// Parse the raw bytes of an image into an Image format.
func parseImg(rawImg []byte) image.Image {
	img, _, err := image.Decode(bytes.NewReader(rawImg))
	if err != nil {
		log.Fatal(err)
	}
	return img
}

// Convert an image to WebP format.
func imgToWebp(img image.Image) WebpImage {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Lossless: true}); err != nil {
		log.Fatal(err)
	}
	return buf
}

// handler for a GET request on `/`
func handleRoot(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		results, err := keikodb.GetHits(db)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("Error getting information from the database!"))
			return
		}
		output := "RESULTS\n"
		for name, count := range results {
			output += fmt.Sprintf("%s: %d\n", name, count)
		}
		w.Write([]byte(output))
	})
}

// handler for a GET request on `/bench`
func handleBench(db *sql.DB, output <-chan *ImageData) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		// start := time.Now()
		imageData := <-output
		// log.Println("--image prepare: time elapsed: ", time.Since(start).Milliseconds())
		log.Println("Serve", imageData.Name)
		w.Write(imageData.WebpImg.Bytes())
		go updateDatabase(db)
	})
}

func getImageChan(out chan<- *ImageData) {
	for {
		imgBase := "data"
		// find all the names of the images
		imgNames := getImageNames(imgBase)
		// pick an index at random
		index := rand.Intn(len(imgNames))
		// get the randomly picked image
		imgEntry := imgNames[index]
		imgName := imgEntry.Name()
		imgData := &ImageData{}
		imagesMutex.Lock()
		_, ok := images[imgName]
		if ok {
			imgData.Name = images[imgName].Name
			imgData.WebpImg = images[imgName].WebpImg
			imgData.Count = images[imgName].Count
		}
		imagesMutex.Unlock()
		if !ok {
			// construct a path to the image in the `data` directory
			imgPath := path.Join(imgBase, imgName)
			imgData = prepareImage(imgPath, imgName)
		} else {
			imgData.Count++
		}
		out <- imgData
		imagesMutex.Lock()
		images[imgName] = imgData
		imagesMutex.Unlock()
	}
}

func prepareImages() {
	imgBase := "data"
	// find all the names of the images
	imgNames := getImageNames(imgBase)
	for _, imgEntry := range imgNames {
		imgName := imgEntry.Name()
		imgPath := path.Join(imgBase, imgName)
		imgData := prepareImage(imgPath, imgName)
		imagesMutex.Lock()
		images[imgName] = imgData
		imagesMutex.Unlock()
	}
}

func prepareImage(imgPath string, imgName string) *ImageData {
	// load the image into memory
	rawImg := loadImg(imgPath)
	// parse into an image.Image
	parsedImg := parseImg(rawImg)
	// convert to WebP
	webpImg := imgToWebp(parsedImg)
	// create image data entry
	return &ImageData{
		Name:    imgName,
		WebpImg: webpImg,
		Count:   0,
	}
}

type ImageData struct {
	Name    string
	WebpImg bytes.Buffer
	Count   int
}

var images = make(map[string]*ImageData)
var imagesMutex = sync.Mutex{}

func updateDatabase(db *sql.DB) {
	// start := time.Now()
	currentHits := make(map[string]int)
	imagesMutex.Lock()
	for k, v := range images {
		currentHits[k] = v.Count
	}
	imagesMutex.Unlock()
	for imgName, count := range currentHits {
		keikodb.SetHitCount(db, imgName, count)
	}
	// log.Println("--database update: time elapsed: ", time.Since(start).Milliseconds())
}

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	log.Println("Using database './data/keiko.db'")
	db, err := sql.Open("sqlite3", "file:./_data/keiko.db?cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	// try to create a new database
	keikodb.MakeNew(db)

	go prepareImages()
	output := make(chan *ImageData)
	go getImageChan(output)

	// serve on `/bench`
	http.Handle("/bench", handleBench(db, output))

	// serve on `/`
	http.Handle("/", handleRoot(db))

	log.Println("Listening on localhost:8099")
	err = http.ListenAndServe("localhost:8099", nil)
	if err != nil {
		log.Fatalln(err)
	}
	close(output)
}
