package browser

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/fatih/color"
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

	log.Println("PROXY :::::::::\n", proxyRes.Header)

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

	log.Println("FINAL RESPONSE ----------\n", w.Header(), url, "...\n ")
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
		cmd := &Cmd{
			Name:      fbBinPath,
			Args:      []string{"config", "init"},
			Alias:     "fbinit",
			ShouldLog: true,
		}
		err = cmd.Exec()
		if err != nil {
			log.Println("Failed to initialize filebrowser configuration")
			log.Fatal(err)
		}
	}

	// filebrowser config set --auth.method=proxy --auth.header=X-Generic-AppName --auth.proxy.showLogin
	cmd := &Cmd{
		Name:      fbBinPath,
		Args:      []string{"config", "set", "--auth.method=proxy", "--auth.header=" + fbAuthHeader, "--auth.proxy.showLogin"},
		Alias:     "fbconf",
		ShouldLog: true,
	}
	err = cmd.Exec()
	if err != nil {
		log.Println("The " + fbBinPath + " might be running, please kill it")
		log.Fatal(err)
	}

	// filebrowser -r storageDir -b /admin
	fbcmd := &Cmd{
		Name:       fbBinPath,
		Args:       []string{"-r", dirname, "-b", fbBaseURL},
		Alias:      "fbrowser",
		Background: true,
		ShouldLog:  true,
	}
	fbcmd.SetLogLevel(log.Lshortfile)
	log.Println("Starting filebrowser...")
	err = fbcmd.Exec()
	if err != nil {
		fbcmd.StderrLogger.Fatal(err)
	}

	fbcmd.Wg.Wait()
}

// Cmd cmd
type Cmd struct {
	// Name the name of the command or the entity point
	Name string
	// Args the command's arguments
	Args []string
	// Alias an alias for the command used when logging
	Alias string
	// Background whether the command should run in background in a goroutine
	//
	// If set to true, you MUST call cmd.Wg.Wait() for program to not exit
	Background bool
	cmd        *exec.Cmd
	Wg         *sync.WaitGroup
	// ShouldLog whether the command should log
	ShouldLog bool
	// The logger used by to log to stdout
	StdoutLogger *log.Logger
	// The logger used by to log to stderr
	StderrLogger *log.Logger
	// DefaultLogLevel log.LstdFlags | log.Lshortfile
	DefaultLogLevel *int
}

// Exec executes a command also syncing the Stdout, stderr to the console
func (c *Cmd) Exec() (err error) {
	c.Wg = &sync.WaitGroup{}
	if c.Name == "" || c.Args == nil {
		return errors.New("Must provide name along with args")
	}
	if c.Alias == "" {
		c.Alias = c.Name
	}
	c.cmd = exec.Command(c.Name, c.Args...)

	if c.ShouldLog {
		err = c.Log()
	}

	if c.Background {
		c.Wg.Add(1)
		go func() {
			err = c.cmd.Wait()
			c.Wg.Done()
		}()
		return err
	}
	err = c.cmd.Wait()
	return err
}

// SetLogLevel sets the default logging level
//
// log.LstdFlags etc..
func (c *Cmd) SetLogLevel(l int) {
	c.DefaultLogLevel = &l
}

// Println prints to stdout
func (c *Cmd) Println(args ...interface{}) {
	c.initLoggers()
	c.StdoutLogger.Println(args...)
}

// Errln prints to sterr
func (c *Cmd) Errln(args ...interface{}) {
	c.initLoggers()
	c.StderrLogger.Println(args...)
}

func (c *Cmd) initLoggers() {
	if c.StderrLogger == nil {
		c.StderrLogger = log.New(os.Stderr, "", *c.DefaultLogLevel)
	}
	if c.StdoutLogger == nil {
		c.StdoutLogger = log.New(os.Stdout, "", *c.DefaultLogLevel)
	}
}

// Log starts logging the command output
func (c *Cmd) Log() error {
	if c.DefaultLogLevel == nil {
		def := log.LstdFlags | log.Lshortfile
		c.DefaultLogLevel = &def
	}
	c.initLoggers()
	out, err := c.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return err
	}
	err = c.cmd.Start()
	if err != nil {
		return err
	}
	outProgress := &stdBuff{}

	outScanner := bufio.NewScanner(out)
	c.StdoutLogger.SetPrefix(color.GreenString("["+c.Alias+"]") + color.CyanString("::stdout "))
	go func(progress *stdBuff) {
		for outScanner.Scan() {
			progress.set(outScanner.Text())
			c.StdoutLogger.Println(progress.get())
			progress.set("")
		}
	}(outProgress)

	// errProgress := &stdBuff{}
	errScanner := bufio.NewScanner(stderr)
	c.StderrLogger.SetPrefix(color.HiRedString("[" + c.Alias + "]" + color.CyanString("::stderr ")))
	go func(progress *stdBuff) {
		for errScanner.Scan() {
			progress.set(errScanner.Text())
			c.StderrLogger.Println(progress.get())
			progress.set("")
		}
	}(outProgress)
	return err
}

// https://stackoverflow.com/questions/20009588/golang-how-to-print-data-from-running-goroutine-at-fixed-intervals#comment29809839_20011457
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
