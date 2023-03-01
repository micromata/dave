// +build mage

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
	"path/filepath"
	"strings"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

const (
	// DIST is the name of the dist directory
	DIST = "dist"
)

type target struct {
	goos   string
	goarch string
}

// Build Builds dave and davecli and moves it to the dist directory
func Build() error {
	mg.Deps(InstallDeps)
	mg.Deps(Clean)

	if _, err := os.Stat(DIST); os.IsNotExist(err) {
		os.Mkdir(DIST, os.ModePerm)
		fmt.Printf("Created dist dir: %s\n", DIST)
	}

	fmt.Println("Building...")

	buildSpecific(target{})

	fmt.Printf("Compiled files moved to folder: %s\n", DIST)

	return nil
}

// BuildReleases Builds dave and davecli for different OS and package them to a zip file for each os
func BuildReleases() error {
	mg.Deps(Clean)

	targets := []target{
		{"windows", "amd64"},
		{"windows", "386"},
		{"darwin", "amd64"},
		{"linux", "amd64"},
		{"linux", "386"},
	}

	for _, t := range targets {
		fmt.Printf("Building for OS %s and architecture %s\n", t.goos, t.goarch)
		dave, daveCli, _ := buildSpecific(t)

		files := []string{
			dave,
			daveCli,
			"Readme.md",
			filepath.Join("examples", "config-sample.yaml"),
		}

		archiveName := fmt.Sprintf("dave-%s-%s.zip", t.goos, t.goarch)
		zipFiles(filepath.Join("dist", archiveName), files)

		os.Remove(dave)
		os.Remove(daveCli)
	}

	return nil
}

// Fmt Formats the code via gofmt
func Fmt() error {
	fmt.Println("Formatting code ...")

	fileList, err := goFileList()
	if err != nil {
		return err
	}

	for _, file := range fileList {
		err = exec.Command("gofmt", "-s", "-l", "-w", file).Run()

		if err != nil {
			return err
		}
	}

	return nil
}

// Check Runs golint and go tool vet on each .go file.
func Check() error {
	fmt.Println("Checking code ...")

	fileList, err := goFileList()
	if err != nil {
		return err
	}

	for _, file := range fileList {
		lintOut, err := exec.Command("golint", file).Output()
		if err != nil {
			return err
		}

		if len(lintOut) > 0 {
			fmt.Println(string(lintOut))
		}

		vetOut, err := exec.Command("go", "tool", "vet", file).Output()
		if len(vetOut) > 0 {
			fmt.Println(string(vetOut))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Install Installs dave and davecli to your $GOPATH/bin folder
func Install() error {
	mg.Deps(InstallDeps)

	fmt.Println("Installing...")
	return exec.Command("go", "install", "./...").Run()
}

// InstallDeps Runs dep ensure and installs additional dependencies.
func InstallDeps() error {
	fmt.Println("Installing Deps...")
	err := exec.Command("dep", "ensure").Run()
	if err != nil {
		return err
	}

	// Install necessary dependency for windows compilation
	return exec.Command("go", "get", "-u", "github.com/inconshreveable/mousetrap").Run()
}

// Clean Removes the dist directory
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(DIST)
}

func goFileList() ([]string, error) {
	fileList := make([]string, 0)
	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".go") {
			fileList = append(fileList, path)
		}

		return err
	})

	return fileList, err
}

func buildSpecific(t target) (string, string, error) {
	env := os.Environ()

	if t.goos != "" && t.goarch != "" {
		env = append(env, fmt.Sprintf("GOOS=%s", t.goos))
		env = append(env, fmt.Sprintf("GOARCH=%s", t.goarch))
	}

	daveSource := filepath.Join("cmd", "dave", "main.go")
	daveExe := filepath.Join(DIST, "dave")
	if t.goos == "windows" {
		daveExe += ".exe"
	}
	daveCommand := exec.Command("go", "build", "-o", daveExe, daveSource)
	daveCommand.Env = env
	err := daveCommand.Run()
	if err != nil {
		return "", "", err
	}

	daveCliSource := filepath.Join("cmd", "davecli", "main.go")
	daveCliExe := filepath.Join(DIST, "davecli")
	if t.goos == "windows" {
		daveCliExe += ".exe"
	}
	daveCliCommand := exec.Command("go", "build", "-o", daveCliExe, daveCliSource)
	daveCliCommand.Env = env
	err = daveCliCommand.Run()
	if err != nil {
		return "", "", err
	}

	return daveExe, daveCliExe, nil
}

// zipFiles compresses one or many files into a single zip archive file.
// The original code was published under MIT licence under https://golangcode.com/create-zip-files-in-go/
func zipFiles(filename string, files []string) error {

	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}
