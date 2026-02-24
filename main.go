package main

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const dbFile = "jars.db"

func main() {
	homedir, _ := os.LookupEnv("HOME")
	jarroot := fmt.Sprintf("%s/.m2/repository", homedir)

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

		txn, txErr := db.Begin()
		if txErr != nil {
			fmt.Fprintf(os.Stderr, "Error starting db transaction: %v\n", txErr)
			os.Exit(1)
		}

		for _, f := range r.File {
			if strings.HasSuffix(f.Name, ".class") {
				_, err := txn.Exec(`INSERT INTO jarclasses(classname, jarfile) VALUES(?, ?)`, f.Name, path)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		txErr = txn.Commit()
		if txErr != nil {
			fmt.Fprintf(os.Stderr, "Error committing db transaction: %v\n", txErr)
			os.Exit(1)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error searching path: %v\n", err)
	}
}
