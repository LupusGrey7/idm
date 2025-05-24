# idm Project

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License: Apache](https://img.shields.io/badge/License-Apache-yellow.svg)](https://opensource.org/licenses/Apache)

### Description
Go Project Structure Learning
A project for learning and experimenting with Go (Golang) a project structure best practices.

## 📖 Project Overview

This repository serves as a learning ground for:
- Proper Go project organization
- Package management with Go Modules
- Common project structure patterns
- Best practices for maintainable Go code
- Testing and CI/CD integration

## ✨ Features
- Example CLI application
- Modular code structure
- Unit test examples
- Linting configuration
- Makefile for common tasks

## 🚀 Getting Started

### Prerequisites
- Go 1.21+ ([installation guide](https://go.dev/doc/install))
- Git
- (Optional) golangci-lint for code analysis
- lintest workflows (codeql, go-test, linters)

### Installation
1. Clone the repository:
```bash
git clone https://github.com/your-username/ibm.git
cd imd
```
### Build Project

```bash
go build -o bin/app ./cmd/app
```

### Running Test
```bash
# Unit tests
go test -v ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting
```bash
# Install golangci-lint (if not installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
golangci-lint run
```

### Docker-compose up
```bash
# go to the project root folder, open the terminal
docker compose up -d
```

### 📄 License
Apache License

### Contacts

