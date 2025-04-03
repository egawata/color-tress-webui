# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands
- Build: `scripts/build.sh` (requires TinyGo)
- Run server: `go run localserver/run_server.go` or `scripts/run_server.sh` (requires goexec)

## Test Commands
- Run all tests: `go test ./...`
- Run package tests: `go test ./tresser`
- Run single test: `go test ./tresser -run TestGetDarkestColor`

## Code Style Guidelines
- Formatting: Run `gofmt -w .` after editing any .go files
- Imports: Standard library first, third-party imports second
- Variable naming: camelCase, short names for RGB values (r, g, b)
- Error handling: Check errors with `if err != nil`, use `fmt.Errorf` for context
- Type definitions: Use struct-based object model with methods
- Test style: Table-driven tests with testify assertions, descriptive names
- Comments: Both English and Japanese comments are acceptable
- Organization: Separate packages for different functionality
