package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Profile            string   `json:"profile"`
	Regions            []string `json:"regions"`
	ExcludedTypes      []string `json:"excluded_types"`
	DryRun             bool     `json:"dry_run"`
	SkipIAMCredentials bool     `json:"skip_iam_credentials"`
}

type Resource struct {
	ID     string
	Type   string
	Region string
}

var safeRegions = []string{"us-east-1", "us-east-2", "us-west-1", "us-west-2", "ap-south-1", "ap-northeast-2", "ap-southeast-1", "ap-northeast-1", "ap-southeast-2", "ca-central-1", "cn-north-1", "cn-northwest-1", "eu-central-1", "eu-west-1", "eu-west-2", "eu-west-3", "eu-north-1", "sa-east-1"}

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
		fmt.Println("v0.2.0")
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Auto AWS Cleanup v0.2.0")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  auto-aws-cleanup init     - Create configuration file")
	fmt.Println("  auto-aws-cleanup login    - Authenticate with AWS")
	fmt.Println("  auto-aws-cleanup list     - List all resources")
	fmt.Println("  auto-aws-cleanup clean    - Clean all resources")
	fmt.Println("  auto-aws-cleanup status   - Show current status")
	fmt.Println("  auto-aws-cleanup version  - Show version")
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
	fmt.Println(string(output))
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
	var totalResources []Resource
	for _, region := range config.Regions {
		fmt.Printf("  📍 Scanning %s...\n", region)
		regionResources := scanRegion(region, config.ExcludedTypes)
		totalResources = append(totalResources, regionResources...)
		fmt.Printf("    Found %d resources\n", len(regionResources))
	}
	fmt.Println("")
	fmt.Printf("📌 Total: %d resources\n", len(totalResources))
	fmt.Println("✓ Scan complete")
}

func scanRegion(region string, excluded []string) []Resource {
	var resources []Resource
	instances := listEC2(region)
	for _, inst := range instances {
		if !isExcluded("EC2", excluded) {
			resources = append(resources, inst)
		}
	}
	stacks := listCloudFormation(region)
	for _, stack := range stacks {
		if !isExcluded("CloudFormation", excluded) {
			resources = append(resources, stack)
		}
	}
	return resources
}

func listEC2(region string) []Resource {
	var resources []Resource
	cmd := exec.Command("aws", "ec2", "describe-instances",
		"--region", region,
		"--query", "Reservations[*].Instances[*].[InstanceId,State.Name]",
		"--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return resources
	}
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		return resources
	}
	if reservations, ok := data["Reservations"].([]interface{}); ok {
		for _, res := range reservations {
			if r, ok := res.(map[string]interface{}); ok {
				if instances, ok := r["Instances"].([]interface{}); ok {
					for _, inst := range instances {
						if i, ok := inst.(map[string]interface{}); ok {
							if id, ok := i["InstanceId"].(string); ok {
								if state, ok := i["State"].(map[string]interface{}); ok {
									if name, ok := state["Name"].(string); ok {
										if name != "terminated" {
											resources = append(resources, Resource{ID: id, Type: "EC2", Region: region})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return resources
}

func listCloudFormation(region string) []Resource {
	var resources []Resource
	cmd := exec.Command("aws", "cloudformation", "describe-stacks",
		"--region", region,
		"--query", "Stacks[*].[StackName,StackStatus]",
		"--output", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return resources
	}
	var stacks []map[string]interface{}
	json.Unmarshal(output, &stacks)
	for _, stack := range stacks {
		if name, ok := stack["StackName"].(string); ok {
			if status, ok := stack["StackStatus"].(string); ok {
				if status != "DELETE_IN_PROGRESS" && status != "DELETE_COMPLETE" {
					resources = append(resources, Resource{ID: name, Type: "CloudFormation", Region: region})
				}
			}
		}
	}
	return resources
}

func isExcluded(t string, excluded []string) bool {
	for _, e := range excluded {
		if e == t {
			return true
		}
	}
	return false
}

func clean() {
	config := loadConfig()
	fmt.Printf("🧹 Starting cleanup (dry-run: %v)...\n", config.DryRun)
	if config.DryRun {
		fmt.Println("🔍 Dry run mode - no resources will be deleted")
		fmt.Println("")
		fmt.Println("To actually delete resources:")
		fmt.Println("  1. Edit .auto-aws-config.json")
		fmt.Println("  2. Set dry_run: false")
		fmt.Println("  3. Run: auto-aws-cleanup clean")
		return
	}
	fmt.Println("⚠️  WARNING: This will permanently delete resources!")
	fmt.Println("")
	fmt.Println("Preserving:")
	for _, t := range config.ExcludedTypes {
		fmt.Printf("  • %s\n", t)
	}
	fmt.Println("")
	deletedCount := 0
	for _, region := range config.Regions {
		fmt.Printf("  📍 Cleaning %s...\n", region)
		count, _ := deleteRegion(region, config.ExcludedTypes)
		deletedCount += count
	}
	fmt.Println("")
	fmt.Printf("✓ Deleted %d resources\n", deletedCount)
}

func deleteRegion(region string, excluded []string) (int, error) {
	var deletedCount int
	instances := listEC2(region)
	for _, inst := range instances {
		if !isExcluded("EC2", excluded) {
			cmd := exec.Command("aws", "ec2", "terminate-instances",
				"--region", region,
				"--instance-ids", inst.ID)
			output, err := cmd.CombinedOutput()
			if err == nil {
				deletedCount++
				fmt.Printf("  🗑️  Terminated EC2: %s\n", inst.ID)
			} else {
				fmt.Printf("  ❌ Failed to terminate %s: %v\n", inst.ID, string(output))
			}
		}
	}
	stacks := listCloudFormation(region)
	for _, stack := range stacks {
		if !isExcluded("CloudFormation", excluded) {
			cmd := exec.Command("aws", "cloudformation", "delete-stack",
				"--region", region,
				"--stack-name", stack.ID)
			output, err := cmd.CombinedOutput()
			if err == nil {
				deletedCount++
				fmt.Printf("  🗑️  Deleted CloudFormation: %s\n", stack.ID)
			} else {
				fmt.Printf("  ❌ Failed to delete %s: %v\n", stack.ID, string(output))
			}
		}
	}
	return deletedCount, nil
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
