package browser

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/asdine/storm"
	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/img"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
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
	}
	err = d.store.Auth.Save(&auth.JSONAuth{})
	checkError(err)

	err = d.store.Settings.Save(set)
	checkError(err)

	// TODO arg
	wd, err := os.Getwd()
	checkError(err)

	ser := &settings.Server{
		BaseURL:          fbBaseURL,
		Port:             serverPort,
		Log:              "stdout",
		Address:          "127.0.0.1",
		Root:             wd,
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
	log.Println(err)
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

	var fileCache diskcache.Interface = diskcache.NewNoOp()
	server, err := d.store.Settings.GetServer()
	checkError(err)

	handler, err := fbhttp.NewHandler(img.New(4), fileCache, d.store, server)
	checkError(err)

	reg := &RegexpHandler{}
	// TODO caching
	reg.Handler(fbBaseURL, handler)
	reg.HandleFunc("/", otherRoutes)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = serverPort
	}
	log.Println("Running on port", PORT)
	err = http.ListenAndServe(":"+PORT, reg)
	checkError(err)
}
