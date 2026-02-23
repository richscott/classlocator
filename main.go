package main

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var classRe = regexp.MustCompile(`\.class$`)

var extractClasses = func(path string, d fs.DirEntry, err error) error {
	if d.IsDir() || !strings.HasSuffix(path, ".jar") {
		return nil
	}

	r, openErr := zip.OpenReader(path)
	if openErr != nil {
		log.Fatal(openErr)
	}
	defer func() {
		closeErr := r.Close()
		if closeErr != nil {
			fmt.Fprintf(os.Stderr, "error closing jar file reader: %v\n", closeErr)
		}
	}()

	fmt.Printf("%s\n", path)
	for _, f := range r.File {
		if classRe.MatchString(f.Name) {
			fmt.Printf("%s\n", f.Name)
		}
	}
	fmt.Printf("\n")
	return nil
}

func main() {
	jarroot := "/Users/richscott/.m2/repository/org/apache/spark/spark-core_2.13/3.5.5"

	err := filepath.WalkDir(jarroot, extractClasses)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching path: %v\n", err)
	}
}
