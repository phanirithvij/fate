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

var (
	target    string
	port      int
	fbBaseURL = "/admin"
)

type pythonData struct {
	hadDB bool
	store *storage.Storage
}

func quickSetup(d *pythonData) {
	k, err := settings.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	set := &settings.Settings{
		AuthMethod: auth.MethodJSONAuth,
		// AuthMethod:    auth.MethodProxyAuth,
		Key:           k,
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
	// err = d.store.Auth.Save(&auth.ProxyAuth{
	// 	Header:        fbAuthHeader,
	// 	ShowLoginPage: true,
	// })
	if err != nil {
		log.Fatal(err)
	}

	err = d.store.Settings.Save(set)
	if err != nil {
		log.Fatal(err)
	}

	wd, _ := os.Getwd()

	ser := &settings.Server{
		BaseURL: fbBaseURL,
		Port:    "8080",
		Log:     "stdout",
		Address: "127.0.0.1",
		Root:    wd,
	}
	err = d.store.Settings.SaveServer(ser)
	if err != nil {
		log.Fatal(err)
	}
	username := "admin"
	password := ""

	if password == "" {
		password, err = users.HashPwd("admin")
		if err != nil {
			log.Fatal(err)
		}
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
	if err != nil {
		log.Println(err)
	}
}

func otherRoutes(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "HELLOE>E>E>>")
}

// StartBrowser starts the filebrowser instance
func StartBrowser(dirname string) {
	log.SetOutput(os.Stdout)
	d := &pythonData{hadDB: true}

	_, err := os.Stat("filebrowser.db")
	log.Println(err)
	if err != nil {
		d.hadDB = false
	}

	db, err := storm.Open("filebrowser.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	d.store, err = bolt.NewStorage(db)
	if err != nil {
		log.Fatal(err)
	}

	if !d.hadDB {
		quickSetup(d)
	}

	var fileCache diskcache.Interface = diskcache.NewNoOp()
	server, err := d.store.Settings.GetServer()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := fbhttp.NewHandler(img.New(4), fileCache, d.store, server)
	if err != nil {
		log.Fatal(err)
	}

	reg := &RegexpHandler{}
	reg.Handler(fbBaseURL, handler)
	reg.HandleFunc("/", otherRoutes)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "3000"
	}
	log.Println("Running on port", PORT)
	err = http.ListenAndServe(":"+PORT, reg)
	if err != nil {
		log.Fatal(err)
	}
}
