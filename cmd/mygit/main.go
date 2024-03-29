package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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

		fmt.Print(stringcontent[6+endIndex:])

	case "hash-object":
		filePath := os.Args[3]
		file, err := os.Open(filePath)
		if err != nil {
			os.Exit(1)
		}
		defer file.Close()
		content, err := os.ReadFile(filePath)
		strContent := string(content)
		if err != nil {
			os.Exit(1)
		}

		hasher := sha1.New()
		if _, err := io.Copy(hasher, file); err != nil {
			os.Exit(1)
		}

		hashByte := hasher.Sum(nil)
		hashString := hex.EncodeToString(hashByte) 


		blobContent := "blob " + strconv.Itoa(len(strContent)) +"\x00"+ strContent

		var compressedData bytes.Buffer

		gz := zlib.NewWriter(&compressedData)

		_, err = gz.Write([]byte(blobContent))
		if err != nil {
			os.Exit(1)
		}

		err = gz.Close()
		if err != nil {
			os.Exit(1)
		}

		fmt.Println(blobContent)
		fmt.Println(compressedData.Bytes())
		fmt.Println(hashString)
		folderame := hashString[:2]
		filename := hashString[2:]
		if err := os.Mkdir(".git/objects/"+ folderame, 0644); err != nil {
			os.Exit(1)
		}
		fmt.Print(filename)
		err = os.WriteFile(".git/objects/"+folderame +"/"+ filename,compressedData.Bytes(), 0644)
		fmt.Print(err)
		if err != nil {
			os.Exit(1)
		}

		
	default: 
		fmt.Fprint(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
