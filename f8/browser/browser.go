// +build !nobrowser

package browser

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/img"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/afero"
)

const (
	// TODO arg
	fbBaseURL  = "/admin"
	fbDBPath   = "filebrowser.db"
	serverPort = "3000"
)

type pythonData struct {
	hadDB bool
	store *storage.Storage
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func quickSetup(d *pythonData) {
	key, err := settings.GenerateKey()
	checkError(err)

	set := &settings.Settings{
		AuthMethod:    auth.MethodJSONAuth,
		Key:           key,
		Signup:        true,
		CreateUserDir: true,
		Defaults: settings.UserDefaults{
			Scope:       ".",
			Locale:      "en",
			SingleClick: false,
			Perm: users.Permissions{
				Admin:    false,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
		},
		Branding: settings.Branding{
			Theme: "dark",
		},
	}
	err = d.store.Auth.Save(&auth.JSONAuth{})
	checkError(err)

	err = d.store.Settings.Save(set)
	checkError(err)

	// TODO arg
	cwd, err := os.Getwd()
	checkError(err)

	ser := &settings.Server{
		BaseURL:          fbBaseURL,
		Port:             serverPort,
		Log:              "stdout",
		Address:          "127.0.0.1",
		Root:             cwd,
		ResizePreview:    false,
		EnableThumbnails: true,
		EnableExec:       false,
	}

	err = d.store.Settings.SaveServer(ser)
	checkError(err)
	username := "admin"
	password := ""

	if password == "" {
		password, err = users.HashPwd("admin")
		checkError(err)
	}

	if username == "" || password == "" {
		log.Fatal("username and password cannot be empty during quick setup")
	}

	user := &users.User{
		Username:     username,
		Password:     password,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	err = d.store.Users.Save(user)
	checkError(err)
}

func otherRoutes(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "HELLOE>E>E>>")
}

// StartBrowser starts the filebrowser instance
func StartBrowser(dirname string) {
	log.SetOutput(os.Stdout)
	d := &pythonData{hadDB: true}

	_, err := os.Stat(fbDBPath)
	if err != nil {
		d.hadDB = false
	}

	db, err := storm.Open(fbDBPath)
	checkError(err)
	defer db.Close()

	d.store, err = bolt.NewStorage(db)
	checkError(err)

	if !d.hadDB {
		quickSetup(d)
	}

	// TODO cache dir
	var fileCache diskcache.Interface = diskcache.NewNoOp()
	cacheDir := "cache"
	if cerr := os.MkdirAll(cacheDir, 0700); cerr != nil {
		log.Fatalf("can't make cache directory %s: %s", cacheDir, cerr)
	}
	fileCache = diskcache.New(afero.NewOsFs(), cacheDir)

	server, err := d.store.Settings.GetServer()
	checkError(err)

	// num cpu cores workers
	numW := maxCPU()
	log.Println("Image processing worker count:", numW)

	handler, err := fbhttp.NewHandler(img.New(numW), fileCache, d.store, server)
	checkError(err)

	reg := &RegexpHandler{}
	fbcHandler := &FBCache{handler: handler}
	reg.Handle(fbBaseURL, fbcHandler)
	reg.HandleFunc("/", otherRoutes)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = serverPort
	}
	log.Println("Running on port", PORT)
	err = http.ListenAndServe(":"+PORT, reg)
	checkError(err)
}

// https://play.golang.org/p/fpETA9_1oo

// FBCache An extension cache handler for filebrowser assets
type FBCache struct {
	handler http.Handler
}

var (
	// server start time
	cacheSince    = time.Now()
	cacheSinceStr = cacheSince.Format(http.TimeFormat)
	cacheUntil    = cacheSince.AddDate(0, 0, 7)
	cacheUntilStr = cacheUntil.Format(http.TimeFormat)
)

// ServeHTTP serves checks if /admin/static then caches && forward or just 304
func (h *FBCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, fbBaseURL+"/static") {
		modtime := r.Header.Get("If-Modified-Since")
		if modtime != "" {
			// TODO check if file is modified since time `t`
			// t, _ := time.Parse(http.TimeFormat, modtime)
			log.Println("[Warning] not checking if modified")
			w.WriteHeader(http.StatusNotModified)
			// no need to forward as cache
			return
		}

		// 604800 -> one week
		w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
		w.Header().Set("Last-Modified", cacheSinceStr)
		w.Header().Set("Expires", cacheUntilStr)
	}
	// forward
	h.handler.ServeHTTP(w, r)
}

// maxCPU max parllelism
func maxCPU() int {
	// https://stackoverflow.com/a/13245047/8608146
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}
