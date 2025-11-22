package startup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// OpenBrowser opens the default browser to the specified URL
func OpenBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	default:
		log.Printf("Unsupported platform: %s. Please open browser manually to: %s", runtime.GOOS, url)
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return err
}

// WaitForServer waits for the server to be ready and then opens the browser
func WaitForServer(port string, timeout time.Duration) {
	url := fmt.Sprintf("http://localhost:%s", port)
	
	// Wait a bit for the server to start
	time.Sleep(2 * time.Second)
	
	// Try to open the browser
	if err := OpenBrowser(url); err != nil {
		log.Printf("Failed to open browser automatically: %v", err)
		log.Printf("Please open your browser and navigate to: %s", url)
	} else {
		log.Printf("🌐 Browser opened successfully to: %s", url)
	}
}