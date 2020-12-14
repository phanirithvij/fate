package main

// https://play.golang.org/p/fpETA9_1oo
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

var oldButtonFormat = regexp.MustCompile("^/([0-9]*)/([0-9]).(jpg|gif|png)$")

var (
	cacheSince = time.Now().AddDate(0, 0, -1)
	// cacheSince    = time.Now()
	cacheSinceStr = cacheSince.Format(http.TimeFormat)
	cacheUntil    = time.Now().AddDate(60, 0, 0)
	cacheUntilStr = cacheUntil.Format(http.TimeFormat)
)

func baseHandler(w http.ResponseWriter, r *http.Request) {
	slices := oldButtonFormat.FindStringSubmatch(r.URL.Path)
	if len(slices) == 4 {
		modtime := r.Header.Get("If-Modified-Since")
		if modtime != "" {
			fmt.Println(modtime)
			// TODO check if file is modified since t
			// t, _ := time.Parse(http.TimeFormat, modtime)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		// Display old images
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		log.Println("last modified", cacheSinceStr)
		log.Println("expires", cacheUntilStr)
		w.Header().Set("Last-Modified", cacheSinceStr)
		w.Header().Set("Expires", cacheUntilStr)

		switch slices[3] {
		case "gif":
			w.Header().Set("Content-Type", "image/gif")
		case "png":
			w.Header().Set("Content-Type", "image/png")
		case "jpg":
			w.Header().Set("Content-Type", "image/jpeg")
		}

		w.Write(oldButtons[slices[3]])
	} else {
		// Display standard paths
		switch r.URL.Path {
		case "/":
			http.Redirect(w, r, "https://comicrank.com", 302)
		case "/robots.txt":
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "User-agent: *\nDisallow: /")
		default:
			http.NotFound(w, r)
		}
	}
}

// Map to hold old button images in memory
var oldButtons map[string][]byte

// Cache given button image's data in memory
func initImg(index string, filename string) {
	file, _ := os.Open(filename)
	info, _ := file.Stat()
	oldButtons[index] = make([]byte, info.Size())
	file.Read(oldButtons[index])
	file.Close()
}

func main() {
	// Cache old buttons in memory
	oldButtons = make(map[string][]byte, 3)
	initImg("gif", "gif.gif")
	initImg("jpg", "jpg.jpg")
	initImg("png", "png.png")

	// Start the http server
	http.HandleFunc("/", baseHandler)
	log.Println("Running on port 8080")
	// eg: request http://localhost:8080/33/3.png
	http.ListenAndServe(":8080", nil)
}
