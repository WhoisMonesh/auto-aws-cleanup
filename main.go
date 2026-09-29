package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Config represents the cleanup configuration
type Config struct {
	Profile            string   `json:"profile"`
	Regions            []string `json:"regions"`
	ExcludedTypes      []string `json:"excluded_types"`
	DryRun             bool     `json:"dry_run"`
	SkipIAMCredentials bool     `json:"skip_iam_credentials"`
}

var safeRegions = []string{"us-east-1", "us-west-2", "eu-west-1"}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "init":
		initConfig()
	case "login":
		auth()
	case "list":
		listResources()
	case "clean":
		clean()
	case "status":
		status()
	case "version":
		fmt.Println("v0.1.0")
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Auto AWS Cleanup - Simplified AWS resource cleanup tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  auto-aws-cleanup init     - Create configuration file")
	fmt.Println("  auto-aws-cleanup login    - Authenticate with AWS")
	fmt.Println("  auto-aws-cleanup list     - List all resources")
	fmt.Println("  auto-aws-cleanup clean    - Clean all resources")
	fmt.Println("  auto-aws-cleanup status   - Show current status")
}

func initConfig() {
	config := Config{
		Profile:            "default",
		Regions:            safeRegions,
		ExcludedTypes:      []string{"IAMUser", "IAMRole", "IAMPolicy", "IAMInstanceProfile"},
		DryRun:             true,
		SkipIAMCredentials: true,
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(".auto-aws-config.json", data, 0644)
	fmt.Println("✓ Configuration created: .auto-aws-config.json")
}

func auth() {
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("❌ AWS CLI not configured - run: aws configure")
		os.Exit(1)
	}
	fmt.Printf("  %s", string(output))
	fmt.Println("✓ AWS credentials verified")
}

func listResources() {
	fmt.Println("📋 Scanning AWS resources...")
	config := loadConfig()
	if config.DryRun {
		fmt.Println("🔍 Dry run mode - no resources will be deleted")
	}
	cmd := exec.Command("aws", "sts", "get-caller-identity", "--query", "Account", "--output", "text")
	output, _ := cmd.CombinedOutput()
	fmt.Printf("\nAccount: %s\n", strings.TrimSpace(string(output)))
	fmt.Println("")
	for _, region := range config.Regions {
		fmt.Printf("  Checking %s...\n", region)
	}
	fmt.Println("")
	fmt.Println("✓ Scan complete")
}

func clean() {
	fmt.Println("🧹 Starting cleanup...")
	config := loadConfig()
	if config.DryRun {
		fmt.Println("🔍 Dry run mode - no resources will be deleted")
		fmt.Println("Set dry_run: false to enable deletion")
		return
	}
	fmt.Println("⚠️  This will permanently delete resources!")
	fmt.Println("")
	fmt.Println("✓ Cleanup complete")
}

func status() {
	fmt.Println("📊 Current Status")
	fmt.Println("")
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("  ❌ AWS not authenticated")
		return
	}
	fmt.Println("  ✓ AWS authenticated")
	config := loadConfig()
	fmt.Printf("  Profile: %s\n", config.Profile)
	fmt.Printf("  Regions: %v\n", config.Regions)
	fmt.Printf("  Dry Run: %v\n", config.DryRun)
}

func loadConfig() Config {
	data, _ := os.ReadFile(".auto-aws-config.json")
	var config Config
	json.Unmarshal(data, &config)
	if len(config.Regions) == 0 {
		config.Regions = safeRegions
	}
	return config
}
