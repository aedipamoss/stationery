package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/aedipamoss/stationery/config"
)

func main() {
	config := config.Config{
		Source: "src",
		Output: "out",
	}
	files, err := ioutil.ReadDir(config.Source)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("files: %v", files)

	fmt.Println("Time to write!")
}
