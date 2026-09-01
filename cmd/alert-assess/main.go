package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zhe-xing/alert-impact-assessment/internal/models"
	"github.com/zhe-xing/alert-impact-assessment/internal/processor"
	"github.com/zhe-xing/alert-impact-assessment/internal/report"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "assess":
		if err := runAssess(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("告警业务影响评估（只读）")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  alert-assess assess --input <file.json> [--output-dir reports] [--mock]")
}

func runAssess(args []string) error {
	input := ""
	outputDir := "reports"
	mock := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--input", "-i":
			if i+1 < len(args) {
				input = args[i+1]
				i++
			}
		case "--output-dir", "-o":
			if i+1 < len(args) {
				outputDir = args[i+1]
				i++
			}
		case "--mock":
			mock = true
		}
	}

	if input == "" {
		return fmt.Errorf("--input is required")
	}

	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}

	alerts, window, err := parseAlerts(data)
	if err != nil {
		return err
	}
	if len(alerts) == 0 {
		return fmt.Errorf("no alerts found in input")
	}

	projectRoot := findProjectRoot()

	processed, err := processor.ProcessAlerts(alerts, processor.ProcessOptions{
		Mock:          mock,
		WindowMinutes: window,
		ProjectRoot:   projectRoot,
	})
	if err != nil {
		return err
	}

	dataSource := "apsarastack_readonly"
	if mock {
		dataSource = "mock"
	}

	paths, err := report.WriteReports(processed, outputDir, dataSource)
	if err != nil {
		return err
	}

	fmt.Println("Reports generated:")
	for name, path := range paths {
		fmt.Printf("  %s: %s\n", name, path)
	}
	return nil
}

func parseAlerts(data []byte) ([]models.AlertInput, int, error) {
	var file models.AlertFile
	if err := json.Unmarshal(data, &file); err != nil {
		var list []models.AlertInput
		if err2 := json.Unmarshal(data, &list); err2 != nil {
			return nil, 0, err
		}
		return list, 30, nil
	}

	window := file.Options.MetricWindowMinutes
	if window <= 0 {
		window = 30
	}
	return file.Alerts, window, nil
}

func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
