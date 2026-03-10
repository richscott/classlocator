package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

const dbFile = "jars.db"

func main() {
	homedir, _ := os.LookupEnv("HOME")
	jarroot := fmt.Sprintf("%s/.m2/repository", homedir)

	cmd := &cli.Command{
		Name:  "classlocator",
		Usage: "search through a hierarchy of jar files and create a sqlite db of classnames to jar file mappings",
		Commands: []*cli.Command{
			{
				Name:  "build",
				Usage: "build",
				Action: func(context.Context, *cli.Command) error {
					return buildDb(jarroot)
				},
			},
			{
				Name:  "search",
				Usage: "search",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					searchClass := cmd.Args().Get(0)
					return searchDb(searchClass)
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
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

func searchDb(className string) error {
	fmt.Printf("searching for %s\n", className)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("error opening %s: %v", dbFile, err)
	}

	sql := fmt.Sprintf(`SELECT classname, jarfile
		FROM jarclasses
		WHERE classname LIKE '%s%%'
		ORDER BY classname ASC, jarfile ASC`, className)

	rows, err := db.Query(sql)
	if err != nil {
		return fmt.Errorf("error querying database: %v", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Fatalf("error: could not close database row cursor: %v", err)
		}
	}()

	var foundClass, foundJar string
	for rows.Next() {
		if err = rows.Scan(&foundClass, &foundJar); err != nil {
			return fmt.Errorf("error scanning result row from database: %v", err)
		}
		fmt.Printf("%s\t%s\n", foundClass, foundJar)
	}

	return nil
}
