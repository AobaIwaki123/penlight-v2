//go:build ignore

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	fmt.Println("=== [Dockerfile & Base Image Integrity Verification] ===")

	dockerfileBytes, err := os.ReadFile("Dockerfile")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read Dockerfile: %v\n", err)
		os.Exit(1)
	}
	dockerfile := string(dockerfileBytes)

	goModBytes, err := os.ReadFile("go.mod")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to read go.mod: %v\n", err)
		os.Exit(1)
	}
	goMod := string(goModBytes)

	hasError := false

	// 1. Verify Go version match between go.mod and Dockerfile
	goModVerRe := regexp.MustCompile(`(?m)^go\s+([0-9]+)\.([0-9]+)`)
	goModMatches := goModVerRe.FindStringSubmatch(goMod)
	if len(goModMatches) < 3 {
		fmt.Fprintf(os.Stderr, "❌ Failed to parse Go version from go.mod\n")
		hasError = true
	} else {
		modMajor, _ := strconv.Atoi(goModMatches[1])
		modMinor, _ := strconv.Atoi(goModMatches[2])

		dockerGoRe := regexp.MustCompile(`(?i)FROM\s+golang:([0-9]+)\.([0-9]+)`)
		dockerGoMatches := dockerGoRe.FindStringSubmatch(dockerfile)
		if len(dockerGoMatches) < 3 {
			fmt.Fprintf(os.Stderr, "❌ Failed to find valid 'FROM golang:X.Y' in Dockerfile\n")
			hasError = true
		} else {
			dockerMajor, _ := strconv.Atoi(dockerGoMatches[1])
			dockerMinor, _ := strconv.Atoi(dockerGoMatches[2])

			if dockerMajor < modMajor || (dockerMajor == modMajor && dockerMinor < modMinor) {
				fmt.Fprintf(os.Stderr, "❌ Go version mismatch: Dockerfile uses golang:%d.%d, but go.mod requires >= %d.%d\n",
					dockerMajor, dockerMinor, modMajor, modMinor)
				hasError = true
			} else {
				fmt.Printf("✅ Go builder version: golang:%d.%d matches go.mod (>= %d.%d)\n",
					dockerMajor, dockerMinor, modMajor, modMinor)
			}
		}
	}

	// 2. Verify Node.js version in Dockerfile (must be Active LTS >= 22)
	nodeVerRe := regexp.MustCompile(`(?i)FROM\s+node:([0-9]+)`)
	nodeMatches := nodeVerRe.FindStringSubmatch(dockerfile)
	if len(nodeMatches) < 2 {
		fmt.Fprintf(os.Stderr, "❌ Failed to find valid 'FROM node:X' in Dockerfile\n")
		hasError = true
	} else {
		nodeMajor, _ := strconv.Atoi(nodeMatches[1])
		if nodeMajor < 22 {
			fmt.Fprintf(os.Stderr, "❌ Outdated Node.js base image: node:%d (Active LTS requires >= 22)\n", nodeMajor)
			hasError = true
		} else {
			fmt.Printf("✅ Node.js builder version: node:%d (Active LTS)\n", nodeMajor)
		}
	}

	// 3. Verify CGO_ENABLED=0 in Go build
	if !strings.Contains(dockerfile, "CGO_ENABLED=0") {
		fmt.Fprintf(os.Stderr, "❌ Dockerfile must specify CGO_ENABLED=0 for static binary compilation\n")
		hasError = true
	} else {
		fmt.Println("✅ Pure Go CGO-free static compilation (CGO_ENABLED=0) confirmed")
	}

	// 4. Verify non-root user execution
	userRe := regexp.MustCompile(`(?m)^USER\s+[^\s]+`)
	if !userRe.MatchString(dockerfile) {
		fmt.Fprintf(os.Stderr, "❌ Dockerfile must define non-root 'USER' directive for container runtime security\n")
		hasError = true
	} else {
		fmt.Println("✅ Non-root user execution confirmed")
	}

	if hasError {
		fmt.Fprintf(os.Stderr, "\n❌ Dockerfile integrity checks failed! Please fix issues before deploying.\n")
		os.Exit(1)
	}

	fmt.Println("✅ All Dockerfile integrity checks passed successfully!")
}
