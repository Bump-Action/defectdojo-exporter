package buildinfo

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var version = flag.Bool("version", false, "Show DefectDojo Exporter version")

// Version must be set via -ldflags '-X'
var Version string

// Init must be called after flag.Parse call.
func Init() {
	if *version {
		printVersion()
		os.Exit(0)
	}
}

func init() {
	oldUsage := flag.Usage
	flag.Usage = func() {
		printVersion()
		oldUsage()
	}
}

func printVersion() {
	_, err := fmt.Fprintf(flag.CommandLine.Output(), "%s\n", Version)
	if err != nil {
		log.Printf("Error output version: %v", err)
	}
}
