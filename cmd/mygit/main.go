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
	"time"
)

type TreeObject struct {
	Mode string
	Name string
	Hash []byte
}

func createSha(content string) string {
	hasher := sha1.New()
	hasher.Write([]byte(content))
	hash := hasher.Sum(nil)
	hexHash := hex.EncodeToString(hash)

	return hexHash
}

func createHistory(sha string, data []byte) {
	folderame := sha[:2]
	filename := sha[2:]

	err := os.RemoveAll(".git/objects/"+ folderame)
    if err != nil {
        os.Exit(1)
    }
	if err := os.Mkdir(".git/objects/"+ folderame, 0644); err != nil {
		os.Exit(1)
	}
	err = os.WriteFile(".git/objects/"+folderame +"/"+ filename, data, 0644)
	if err != nil {
		os.Exit(1)
	}
}

func compressData(blobContent string) bytes.Buffer {
	var compressedData bytes.Buffer

	gz := zlib.NewWriter(&compressedData)

	_, err := gz.Write([]byte(blobContent))
	if err != nil {
		os.Exit(1)
	}

	err = gz.Close()
	if err != nil {
		os.Exit(1)
	}

	return compressedData
}

func createBlob(filePath string) string {
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

	blobContent := "blob " + strconv.Itoa(len(strContent)) +"\x00"+ strContent


	sha := createSha(blobContent)
	compressedData := compressData(blobContent)



	createHistory(sha, compressedData.Bytes())

	return sha
}

func createBlobWithContent( content string) string {


	sha := createSha(content)
	compressedData := compressData(content)



	createHistory(sha, compressedData.Bytes())

	return sha
}

func createTree(directory string) string {
	files, err := os.ReadDir(directory)
	if err != nil {
		os.Exit(1)
	}
	content := ""
	for _, item := range files {
		if (item.Name() == ".git") {
			continue;
		}
		if(item.IsDir()) {
			content += "40000" + " " + item.Name() + "\x00"
			sha := createTree(filepath.Join(directory,item.Name()))
			binaryHash, err := hex.DecodeString(sha)
			if err != nil {
				fmt.Println("Error decoding hexadecimal:", err)
			}
			content += string(binaryHash)
		} else {
			content += "100644" + " " + item.Name() + "\x00"
			sha := createBlob(filepath.Join(directory,item.Name()))
			binaryHash, err := hex.DecodeString(sha)
			if err != nil {
				fmt.Println("Error decoding hexadecimal:", err)
			}
			content += string(binaryHash)
		} 
	}

	output := "tree "+ strconv.Itoa(len(content)) +"\x00"+ content
	sha := createSha(output)
	compressedData := compressData(output)

	createHistory(sha, compressedData.Bytes())

	return sha
}



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
			fmt.Print(err)
			os.Exit(1)
		}


		gzread, err := zlib.NewReader(bytes.NewReader(file))
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
		r,err := io.ReadAll(gzread)
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		stringcontent := string(r)
		endIndex := strings.Index(stringcontent[5:], "\000")

		fmt.Print(stringcontent[6+endIndex:])

	case "ls-tree":
		hash := os.Args[3]
		path := filepath.Join(".git", "objects", hash[:2], hash[2:])
		file, err := os.ReadFile(path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}


		gzread, err := zlib.NewReader(bytes.NewReader(file))
		if err != nil {
			fmt.Println("aaabbb")
			os.Exit(1)
		}


		r,err := io.ReadAll(gzread)
		if err != nil {
			fmt.Println("aaabccb")
			os.Exit(1)
		}


		stringcontent := r
		endIndex := bytes.IndexByte(stringcontent, '\x00')

		tree := stringcontent[endIndex+1:]

		for len(tree) > 0 {
			space := bytes.IndexByte(tree, ' ')
			nullbyte := bytes.IndexByte(tree, '\x00')
			neww := TreeObject {
				Mode: string(tree[:space]),
				Name: string(tree[space+1: nullbyte]),
				Hash: tree[nullbyte +1: nullbyte+1 + 20],
			}
			tree = tree[nullbyte+1 + 20:]
			fmt.Println(neww.Name)
		}

	case "write-tree":
		currentDic, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		sha := createTree(currentDic)
		fmt.Println(sha)


	

	case "commit-tree":
		treeSha := os.Args[2]
		commitSha := os.Args[4]
		message := os.Args[6]
		output:= ""
		output += "tree " + treeSha + "\n"
		if(commitSha != "") {
			output += "parent " + commitSha + "\n"
		}
		cTime := time.Now().Unix()
		output += "author abcd <abcd@gmail.com> " +  strconv.FormatInt(cTime, 10) + " -070 \n" 
		output += "commiter abcd <abcd@gmail.com> " + strconv.FormatInt(cTime, 10) + " -0700 \n" 
		output += "\n" + message + "\n"



		commitHeader := "commit "+ strconv.Itoa(len(output)) +"\x00"+ output

		fmt.Println(commitHeader);
		sha := createSha(commitHeader)

		compressedData := compressData(commitHeader)
		
		createHistory(sha, compressedData.Bytes())
		fmt.Print(sha)

	case "hash-object":
		filePath := os.Args[3]
		sha := createBlob(filePath)
		fmt.Print(sha)

		
	default: 
		fmt.Fprint(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}



			// hasher := sha1.New()
			// hasher.Write([]byte(blobContent))
			// hash := hasher.Sum(nil)
			// hexHash := hex.EncodeToString(hash)