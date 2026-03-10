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

	err := buildDb(jarroot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func buildDb(jarroot string) error {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("error opening %s: %v", dbFile, err)
	}

	if _, err = db.Exec(`
		drop table if exists jarclasses;
		create table jarclasses(classname text, jarfile text);
		`); err != nil {
		return fmt.Errorf("error creating database table: %v", err)
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
			log.Fatalf("Error starting db transaction: %v\n", txErr)
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
			log.Fatalf("Error committing db transaction: %v\n", txErr)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error searching path: %v", err)
	}

	return nil
}
