package browser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	target       string
	port         int
	fbBaseURL    = "/admin"
	fbAuthHeader = `X-Generic-AppName`
	fbBinPath    = "filebrowser-custom"
	upgrader     = websocket.Upgrader{} // use default options
)

// Forwarding ...
//
// Got from https://gist.github.com/phanirithvij/24c2700cdcff3d73b7288b0ca265c04b
func Forwarding() {
	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	proxy, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	go copyIO(conn, proxy)
	go copyIO(proxy, conn)
}

func copyIO(src, dest net.Conn) {
	defer src.Close()
	defer dest.Close()
	io.Copy(dest, src)
}

// func allRoutes(w http.ResponseWriter, req *http.Request) {
// 	for name, headers := range req.Header {
// 		for _, h := range headers {
// 			fmt.Fprintf(w, "%v: %v\n", name, h)
// 		}
// 	}
// }

// User ...
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func fileBrowser(w http.ResponseWriter, req *http.Request) {
	url := req.URL
	// TODO http works fine because it is running locally
	url.Scheme = "http"
	url.Host = "localhost:8080"

	// https://stackoverflow.com/a/34724481/8608146
	proxyReq, err := http.NewRequest(req.Method, url.String(), req.Body)
	if err != nil {
		// handle error
		log.Panic(err)
	}

	// clone before doing anything to them
	proxyReq.Header = req.Header.Clone()
	proxyReq.Header.Set("Host", req.Host)
	proxyReq.Header.Set("X-Forwarded-For", req.RemoteAddr)

	// Custom auth for the user
	// https://filebrowser.org/configuration/authentication-method#proxy-header
	if req.Method == "POST" && strings.Contains(url.Path, "/login") {
		// login do our custom login
		var us User
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&us)
		if err != nil {
			log.Panic(err)
		}
		// We've got the username and password
		// log.Println(us.Username, us.Password)
		log.Println(us)
		// now we need to check if such user exists in the server database
		// if found set a header `X-Generic-AppName` with username is allowed

		// TODO query the users from the postgers database
		foundIndDB := true
		if foundIndDB {
			proxyReq.Header.Set(fbAuthHeader, us.Username)
		}
	}

	// ws://.. shell commands
	if strings.Contains(url.Path, "api/command") {
		url.Scheme = "ws"
		clientC, err := upgrader.Upgrade(w, req, nil)
		// clientC, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Println("upgrade:", err)
			return
		}

		fbC, resp, err := websocket.DefaultDialer.Dial(url.String(), nil)
		if err != nil {
			log.Println(fbC, resp)
			log.Fatal("dial:", err)
		}

		err = clientC.WriteMessage(websocket.TextMessage, []byte("message"))
		if err != nil {
			log.Println("write:", err)
			return
		}
		log.Println("sent message:")

		// errChan := make(chan error, 6)
		// // done := make(chan bool, 4)
		// cp := func(dst *websocket.Conn, src *websocket.Conn) {
		// 	defer func() {
		// 		log.Println("Defer cp empty pass")
		// 		errChan <- errors.New("")
		// 	}()
		// 	for {
		// 		mt, message, err := src.ReadMessage()
		// 		if err != nil {
		// 			log.Println("read:", err)
		// 			errChan <- err
		// 			return
		// 		}
		// 		log.Printf("recv: %s", message)
		// 		err = fbC.WriteMessage(mt, message)
		// 		if err != nil {
		// 			log.Println("write:", err)
		// 			errChan <- err
		// 			return
		// 		}
		// 		log.Printf("send: %s", message)
		// 	}
		// }

		// // Start proxying websocket data
		// go cp(fbC, clientC)
		// go cp(clientC, fbC)
		// // TODO why not work ma god
		// <-errChan
		// log.Println("Returning...")
		return
	}

	client := &http.Client{}

	proxyRes, err := client.Do(proxyReq)
	if err != nil {
		log.Panic(err)
	}
	defer proxyRes.Body.Close()

	// Copy code
	w.WriteHeader(proxyRes.StatusCode)

	log.Println(proxyRes.Header.Get("Content-Type"))

	uparts := strings.Split(url.String(), ".")
	ext := "." + uparts[len(uparts)-1]
	// Copy headers as no clone method for function, no lvalues :(
	// fmt.Println("\nheader, values.............", url)
	for header, values := range proxyRes.Header.Clone() {
		for _, value := range values {
			if (header == "Content-Type") && mime.TypeByExtension(ext) == "text/css" {
				w.Header().Set(header, value)
				break
			}
			w.Header().Add(header, value)
		}
	}

	// Copy body
	io.Copy(w, proxyRes.Body)

	log.Println(w.Header(), url)
}

// StartBrowser starts the filebrowser instance
func StartBrowser(dirname string) {
	// go Forwarding()
	go func() {
		reg := &RegexpHandler{}
		reg.HandleFunc(fbBaseURL+"/*", fileBrowser)
		// reg.HandleFunc("/", allRoutes)
		PORT := os.Getenv("PORT")
		if PORT == "" {
			PORT = "3000"
		}
		log.Println("Running on port", PORT)
		err := http.ListenAndServe(":"+PORT, reg)
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err := os.Stat("filebrowser.db")
	if err != nil {
		// need to do this first
		err = Exec(fbBinPath, "config", "init")
		if err != nil {
			log.Println("Failed to initialize filebrowser configuration")
			log.Fatal(err)
		}
	}

	// filebrowser config set --auth.method=proxy --auth.header=X-Generic-AppName --auth.proxy.showLogin
	// cmd := exec.Command("filebrowser", "config", "cat")
	err = Exec(fbBinPath, "config", "set", "--auth.method=proxy", "--auth.header="+fbAuthHeader, "--auth.proxy.showLogin")
	if err != nil {
		log.Println("The " + fbBinPath + " might be running, please kill it")
		log.Fatal(err)
	}

	log.Println("Starting filebrowser...")

	// filebrowser -r storageDir -b /admin
	err = Exec(fbBinPath, "-r", dirname, "-b", fbBaseURL)
	if err != nil {
		log.Fatal(err)
	}
}

// Exec executes a command also syncing the Stdout, stderr to the console
func Exec(name string, arg ...string) (err error) {
	cmd := exec.Command(name, arg...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	outProgress := &stdBuff{}

	outScanner := bufio.NewScanner(out)
	go func(progress *stdBuff) {
		for outScanner.Scan() {
			progress.set(fmt.Sprintln("["+name+"]::stdout", outScanner.Text()))
			fmt.Fprint(os.Stdout, progress.get())
		}
	}(outProgress)

	errProgress := &stdBuff{}
	errScanner := bufio.NewScanner(stderr)
	go func(progress *stdBuff) {
		for errScanner.Scan() {
			progress.set(fmt.Sprintln("["+name+"]::stderr", errScanner.Text()))
			fmt.Fprint(os.Stderr, progress.get())
		}
	}(errProgress)

	err = cmd.Wait()
	return err
}

type stdBuff struct {
	sync.RWMutex
	current string
}

func (p *stdBuff) set(value string) {
	p.Lock()
	defer p.Unlock()
	p.current = value
}

func (p *stdBuff) get() string {
	p.RLock()
	defer p.RUnlock()
	return p.current
}
