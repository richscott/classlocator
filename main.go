package main

import (
	"archive/zip"
	"fmt"
	"log"
)

func main() {
	jarfile := "/Users/richscott/.m2/repository/org/apache/spark/spark-core_2.13/3.5.5/spark-core_2.13-3.5.5.jar"

	// Open a zip archive for reading.
	r, err := zip.OpenReader(jarfile)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		fmt.Printf("%s\n", f.Name)
	}
}
