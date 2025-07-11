# Micro-Go Project

## Project Overview
TinyGo/LLVM microcontroller project targeting Xtensa/ESP32 architecture.

## Project Structure
- `env.sh`: Environment setup with LLVM build paths and CGO flags for TinyGo compilation
- `go.mod`: Go module with TinyGo LLVM bindings dependency  
- `main.go`: LLVM IR generator that creates a simple "hello" program for ESP32

## Environment Setup
Before running the project, source the environment file:
```bash
source env.sh
```

## Build/Run Commands
```bash
# Run with byollvm tag
go run -tags byollvm main.go
```

## Dependencies
- golang.org/x/tools v0.34.0
- tinygo.org/x/go-llvm v0.0.0-20250422114502-b8f170971e74

## Target Platform
- Architecture: Xtensa
- Target: ESP32
- Target Triple: "xtensa"