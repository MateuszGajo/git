package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// hash := "95d09f2b10159347eece71399a7e2e907ea3df4f"

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprint(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		hash := os.Args[3]
		path := filepath.Join(".git", "objects", hash[:2], hash[2:])

		file, err := os.ReadFile(path)
		if err != nil {
			os.Exit(1)
		}

		gzread, err := zlib.NewReader(bytes.NewReader(file))
		if err != nil {
			os.Exit(1)
		}
		r,err := io.ReadAll(gzread)
		if err != nil {
			os.Exit(1)
		}

		stringcontent := string(r)
		endIndex := strings.Index(stringcontent[5:], "\000")

		fmt.Println(stringcontent[5+endIndex:])
	default: 
		fmt.Fprint(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
