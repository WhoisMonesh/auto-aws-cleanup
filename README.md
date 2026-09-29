# Auto AWS Cleanup

A simplified AWS resource cleanup tool that automates authentication and resource cleanup. Built in Go with multi-platform support.

## Features

- **Simple CLI**: Easy-to-use commands instead of complex YAML configs
- **Safe by Default**: Dry-run mode enabled by default
- **Multi-Platform**: Works on macOS, Linux, and Windows
- **Automatic Authentication**: Verifies AWS credentials automatically
- **IAM Protection**: Preserves IAM users and access keys by default

## Quick Start

### 1. Installation

Download a binary for your platform from [Releases](https://github.com/whoism/auto-aws-cleanup/releases) or build from source:

```bash
make build
```

### 2. Configure AWS

```bash
aws configure
```

### 3. Initialize Auto AWS Cleanup

```bash
./auto-aws-cleanup init
```

This creates `.auto-aws-config.json` with safe defaults.

### 4. Authenticate

```bash
./auto-aws-cleanup login
```

### 5. List Resources

```bash
./auto-aws-cleanup list
```

### 6. Clean Resources

```bash
# Dry run first (default)
./auto-aws-cleanup clean

# Actual cleanup (edit config to set dry_run: false)
./auto-aws-cleanup clean
```

## Configuration

Edit `.auto-aws-config.json`:

```json
{
  "profile": "default",
  "regions": ["us-east-1", "us-west-2", "eu-west-1"],
  "excluded_types": ["IAMUser", "IAMRole", "IAMPolicy"],
  "dry_run": true,
  "skip_iam_credentials": true
}
```

## Commands

| Command   | Description                          |
|-----------|--------------------------------------|
| `init`    | Create configuration file            |
| `login`   | Authenticate with AWS                |
| `list`    | List all resources                   |
| `clean`   | Clean all resources                  |
| `status`  | Show current status                  |

## Building

```bash
# Build for current platform
make build

# Build for all platforms
make all

# Clean build artifacts
make clean
```

## Configuration Reference

```json
{
  "profile": "default",
  "regions": ["us-east-1", "us-west-2"],
  "excluded_types": ["IAMUser", "IAMRole", "IAMPolicy"],
  "dry_run": true,
  "skip_iam_credentials": true
}
```

| Field | Description |
|-------|-------------|
| `profile` | AWS CLI profile name |
| `regions` | AWS regions to scan/clean |
| `excluded_types` | Resource types to preserve |
| `dry_run` | Test mode (default: true) |
| `skip_iam_credentials` | Preserve IAM users/keys (default: true) |

## License

MIT License