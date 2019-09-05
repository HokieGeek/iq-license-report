package main

import (
	// "fmt"
	"bufio"
	"flag"
	"fmt"
	"os"

	nexusiq "github.com/sonatype-nexus-community/gonexus/iq"
)

func main() {
	// iqServerPtr := flag.String("iq", "http://localhost:8070", "")
	// iqAuthPtr := flag.String("iqAuth", "admin:admin123", "")

	appIDPtr := flag.String("--appId", "Bitwarden-core", "The public ID of the app whose report you need")
	stagePtr := flag.String("--stage", "build", "The stage of the report for the given app")

	flag.Parse()

	iq, err := nexusiq.New("http://localhost:8070", "admin", "admin123")
	if err != nil {
		panic(err)
	}

	report, err := nexusiq.GetRawReportByAppID(iq, *appIDPtr, *stagePtr)
	if err != nil {
		panic(err)
	}

	fmt.Println(report.Components[0].LicensesData)
	// open output file
	fo, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	w := bufio.NewWriter(fo)

	iqlicencereport.WriteHTML(w, nil)
}
