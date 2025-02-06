// checkmd5 is used to compare the md5 hash of a file to a specified value.
package main

import (
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
)

var Version string = "0.1.1"

const Usage = `checkmd5 [--hash=<some hash>] --file=<some file>`

func main() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	getVersion := flag.Bool("version", false,
		"display the version")

	getHelp := flag.Bool("help", false,
		"display the command usage")

	cmpHash := flag.String("hash", "",
		"the md5 hash to compare")

	cmpFile := flag.String("file", "",
		"the file to check")

	flag.Parse()

	if *getVersion {
		fmt.Println("checkmd5", Version)

		os.Exit(0)

		return
	}

	if *cmpFile == "" || *getHelp {
		fmt.Println("Usage:", Usage)

		os.Exit(1)

		return
	}

	vb, err := os.ReadFile(*cmpFile)
	if err != nil {
		fmt.Println("Error reading file:", err.Error())

		os.Exit(1)

		return
	}

	h := HashPackageContents(vb)

	hh := hex.EncodeToString(h)

	if *cmpHash != hh {
		fmt.Println("Provided hash:", *cmpHash,
			", does not match file hash:", hh)

		os.Exit(1)

		return
	}

	os.Exit(0)
}

// HashPackageContents returns a consistent hash for package file contents.
func HashPackageContents(contents []byte) []byte {
	h := sha512.Sum512(contents)

	return h[:]
}
