package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	// "regexp"
	"strings"

	_ "modernc.org/sqlite"
)

const jarroot = "/Users/richscott/.m2/repository"
const dbFile = "jars.db"

// var classRe = regexp.MustCompile(`\.class$`)

func main() {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", dbFile, err)
		os.Exit(1)
	}

	if _, err = db.Exec(`
		drop table if exists jarclasses;
		create table jarclasses(classname text, jarfile text);
		`); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating database table: %v\n", err)
		os.Exit(1)
	}

	err = filepath.WalkDir(jarroot, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || !strings.HasSuffix(path, ".jar") {
			return nil
		}

		fmt.Printf("%s\n", path)

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

		for _, f := range r.File {
			// if classRe.MatchString(f.Name) {
			if strings.HasSuffix(f.Name, ".class") {
				// fmt.Printf("%s\n", f.Name)
				_, err := db.Exec(`INSERT INTO jarclasses(classname, jarfile) VALUES(?, ?)`, f.Name, path)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching path: %v\n", err)
	}
}
