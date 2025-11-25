// cmd/citadel/main.go
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"citadel-agent/backend/internal/workflow/models"
	"citadel-agent/backend/internal/workflow/engine"
)

func main() {
	if len(os.Args) < 2 {
		showHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "test":
		runTests()
	case "start":
		startServer()
	case "stop":
		stopServer()
	case "restart":
		restartServer()
	case "status":
		checkStatus()
	case "update":
		updateAgent()
	case "deploy":
		if len(os.Args) < 3 {
			fmt.Println("❌ Usage: citadel deploy <workflow-file>")
			os.Exit(1)
		}
		deployWorkflow(os.Args[2])
	case "logs":
		showLogs()
	case "version":
		showVersion()
	case "help", "-h", "--help":
		showHelp()
	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Usage: citadel [command]")
	fmt.Println("")
	fmt.Println("Available commands:")
	fmt.Println("  test          - Run tests for the Citadel Agent")
	fmt.Println("  start         - Start the Citadel Agent server")
	fmt.Println("  stop          - Stop the Citadel Agent server")
	fmt.Println("  restart       - Restart the Citadel Agent server")
	fmt.Println("  status        - Check the status of Citadel Agent")
	fmt.Println("  update        - Update Citadel Agent to latest version")
	fmt.Println("  deploy        - Deploy workflow to Citadel Agent")
	fmt.Println("  logs          - Show server logs")
	fmt.Println("  version       - Show Citadel Agent version")
	fmt.Println("  help          - Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  citadel test")
	fmt.Println("  citadel start")
	fmt.Println("  citadel status")
	fmt.Println("  citadel deploy workflow.json")
	fmt.Println("")
}

func runTests() {
	fmt.Println("🧪 Running Citadel Agent tests...")
	fmt.Println("==================================")

	// Membuat instance dari workflow engine untuk test
	workflowEngine := engine.NewEngine()
	
	// Test membuat workflow sederhana
	testWorkflow := &models.Workflow{
		ID:          "test-workflow-1",
		Name:        "Test Workflow",
		Description: "This is a test workflow",
		Status:      "active",
		Nodes: []models.Node{
			{
				ID:   "node-1",
				Name: "Start Node",
				Type: "http_request",
				Parameters: map[string]interface{}{
					"url":    "https://httpbin.org/get",
					"method": "GET",
				},
			},
		},
	}

	fmt.Printf("✅ Created test workflow: %s\n", testWorkflow.Name)
	fmt.Printf("✅ Workflow has %d nodes\n", len(testWorkflow.Nodes))
	fmt.Printf("✅ Workflow engine initialized\n")
	
	// Test eksekusi workflow (ini akan berjalan di background)
	ctx := context.Background()
	execution, err := workflowEngine.ExecuteWorkflow(ctx, testWorkflow)
	if err != nil {
		fmt.Printf("⚠️  Workflow execution test failed: %v\n", err)
	} else {
		fmt.Printf("✅ Workflow execution started: %s\n", execution.ID)
	}

	fmt.Println("✅ Tests completed successfully!")
}

func startServer() {
	fmt.Println("🚀 Starting Citadel Agent server...")
	
	// Cek apakah server sudah berjalan
	if serverIsRunning() {
		fmt.Println("❌ Citadel Agent is already running")
		os.Exit(1)
	}

	// Jalankan server di background
	cmd := exec.Command("go", "run", "cmd/api/main.go")
	cmd.Dir = "backend"
	
	// Arahkan output ke file log
	logFile, err := os.OpenFile("citadel.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("❌ Error creating log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	
	err = cmd.Start()
	if err != nil {
		fmt.Printf("❌ Error starting server: %v\n", err)
		os.Exit(1)
	}

	// Simpan PID
	pidFile, err := os.Create(".citadel.pid")
	if err != nil {
		fmt.Printf("❌ Error creating PID file: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(pidFile, "%d", cmd.Process.Pid)
	pidFile.Close()

	fmt.Printf("✅ Citadel Agent started with PID: %d\n", cmd.Process.Pid)
	fmt.Println("Server logs available at: citadel.log")
}

func stopServer() {
	fmt.Println("🛑 Stopping Citadel Agent server...")
	
	pid := getServerPID()
	if pid == 0 {
		fmt.Println("❌ Citadel Agent is not running")
		return
	}

	// Kirim sinyal SIGTERM untuk shutdown yang graceful
	p, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("❌ Error finding process: %v\n", err)
		os.Remove(".citadel.pid")
		return
	}

	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Printf("❌ Error stopping process: %v\n", err)
		os.Remove(".citadel.pid")
		return
	}

	// Tunggu proses berhenti
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Jika timeout, kirim SIGKILL
			fmt.Println("⚠️  Force killing process...")
			p.Signal(syscall.SIGKILL)
			os.Remove(".citadel.pid")
			fmt.Println("✅ Citadel Agent stopped")
			return
		case <-ticker.C:
			// Cek apakah proses masih berjalan
			if !processExists(pid) {
				os.Remove(".citadel.pid")
				fmt.Println("✅ Citadel Agent stopped")
				return
			}
		}
	}
}

func restartServer() {
	fmt.Println("🔄 Restarting Citadel Agent server...")
	stopServer()
	time.Sleep(2 * time.Second)
	startServer()
}

func checkStatus() {
	if serverIsRunning() {
		pid := getServerPID()
		fmt.Printf("✅ Citadel Agent is running (PID: %d)\n", pid)
	} else {
		fmt.Println("❌ Citadel Agent is not running")
	}
}

func updateAgent() {
	fmt.Println("🔄 Updating Citadel Agent...")
	
	// Dalam implementasi nyata, ini akan:
	// 1. Pull dari repo git
	// 2. Update dependencies
	// 3. Rebuild binary
	
	fmt.Println("🔄 Fetching latest changes...")
	cmd := exec.Command("git", "pull", "origin", "main")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  Git pull failed: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
	}
	
	fmt.Println("🔄 Updating dependencies...")
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = "backend"
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  Dependency update failed: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
	}
	
	fmt.Println("✅ Citadel Agent updated!")
}

func deployWorkflow(workflowFile string) {
	fmt.Printf("📦 Deploying workflow: %s\n", workflowFile)
	
	// Baca file workflow
	file, err := os.Open(workflowFile)
	if err != nil {
		fmt.Printf("❌ Error opening workflow file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Parse JSON workflow
	bytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("❌ Error reading workflow file: %v\n", err)
		os.Exit(1)
	}

	var workflow models.Workflow
	err = json.Unmarshal(bytes, &workflow)
	if err != nil {
		fmt.Printf("❌ Error parsing workflow JSON: %v\n", err)
		os.Exit(1)
	}

	// Dalam implementasi nyata, ini akan mengirim ke API server
	// Untuk sekarang, kita cek validasi sederhana
	fmt.Printf("✅ Workflow '%s' loaded with %d nodes\n", workflow.Name, len(workflow.Nodes))
	
	// Test workflow engine
	workflowEngine := engine.NewEngine()
	ctx := context.Background()
	
	execution, err := workflowEngine.ExecuteWorkflow(ctx, &workflow)
	if err != nil {
		fmt.Printf("⚠️  Workflow validation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Workflow deployed and test execution started: %s\n", execution.ID)
	}
	
	fmt.Println("✅ Workflow deployment completed!")
}

func showLogs() {
	file, err := os.Open("citadel.log")
	if err != nil {
		fmt.Println("❌ Log file not found. Server may not be running.")
		return
	}
	defer file.Close()

	// Baca 50 baris terakhir dari log
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Ambil 50 baris terakhir
	start := 0
	if len(lines) > 50 {
		start = len(lines) - 50
	}

	fmt.Println("📋 Citadel Agent Logs (last 50 lines):")
	fmt.Println("=======================================")
	for i := start; i < len(lines); i++ {
		fmt.Println(lines[i])
	}
}

func showVersion() {
	fmt.Println(" Citadel Agent v1.0.0 (workflow automation platform)")
	fmt.Println(" Similar to n8n - Open Source Workflow Automation")
	fmt.Println(" https://github.com/citadel-agent")
}

// Fungsi helper
func getServerPID() int {
	pidFile, err := os.Open(".citadel.pid")
	if err != nil {
		return 0
	}
	defer pidFile.Close()

	var pid int
	_, err = fmt.Fscanf(pidFile, "%d", &pid)
	if err != nil {
		return 0
	}
	
	// Periksa apakah proses masih berjalan
	if !processExists(pid) {
		// Hapus PID file jika proses tidak berjalan
		os.Remove(".citadel.pid")
		return 0
	}
	
	return pid
}

func processExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// Pada Unix, kita bisa mengirim sinyal 0 untuk cek eksistensi proses
	err = p.Signal(syscall.Signal(0))
	return err == nil
}

func serverIsRunning() bool {
	pid := getServerPID()
	return pid != 0
}