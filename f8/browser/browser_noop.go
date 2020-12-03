// +build nobrowser

package browser

import (
	"log"
)

// StartBrowser does nothing
func StartBrowser(dirname string) {
	// log.SetOutput(os.Stdout)
	log.Println("[Warning] Browser module is not available")
}
