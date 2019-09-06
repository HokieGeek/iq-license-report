package main

import (
	"bufio"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	nexusiq "github.com/sonatype-nexus-community/gonexus/iq"
)

func writeHTML(w io.Writer, appID string, licenses []nexusiq.License) error {
	page := struct {
		Title, Created string
		Licenses       []nexusiq.License
	}{
		Title:    fmt.Sprintf("Licenses for %s", appID),
		Licenses: licenses,
		Created:  time.Now().Format(time.RFC1123),
	}

	t := template.Must(template.New("licenses").Parse(`
	<html>
		<head>
			<title>{{ .Title }}</title>
		</head>
		<body>
			<h1>{{ .Title }}</h1>
			<ul>
			{{ range .Licenses }}
				<li>{{ .LicenseName }}</li>
			{{ end }}
			</ul>
			<footer>Created {{ .Created }} </footer>
		</body>
	</html>`))
	if err := t.Execute(w, page); err != nil {
		return fmt.Errorf("could not generate page: %v", err)
	}

	return nil
}

func main() {
	iqServerPtr := flag.String("iq", "http://localhost:8070", "The host and port of the IQ Server with the report")
	iqAuthPtr := flag.String("auth", "admin:admin123", "The username and password to use with the IQ Server")

	appIDPtr := flag.String("appId", "", "The public ID of the app whose report you need")
	stagePtr := flag.String("stage", "build", "The stage of the report for the given app")

	filePtr := flag.String("file", "", "Filename to save the HTML report as")
	servePortPtr := flag.String("serve", "", "Set the value a port to serve the HTML report on")

	flag.Parse()

	if *appIDPtr == "" {
		panic("appId is a required argument")
	}

	// Create gonexus IQ client
	auth := strings.Split(*iqAuthPtr, ":")
	iq, err := nexusiq.New(*iqServerPtr, auth[0], auth[1])
	if err != nil {
		panic(err)
	}

	// Retrieve the report struct based on appID and scan stage
	report, err := nexusiq.GetRawReportByAppID(iq, *appIDPtr, *stagePtr)
	if err != nil {
		panic(err)
	}

	// Creates a string set of the declared and observed licenses from the report
	// Declared License: these are the licenses that the developer of the component has identified.
	// Observed License: these are the licenses that have been observed during Sonatype’s research.
	licensesSet := make(map[nexusiq.License]struct{})
	for _, c := range report.Components {
		for _, l := range c.LicensesData.DeclaredLicenses {
			licensesSet[l] = struct{}{}
		}
		for _, l := range c.LicensesData.ObservedLicenses {
			licensesSet[l] = struct{}{}
		}
	}

	// Creates a slice ouf of the string set, and filters out noisy "licenses"
	licenses := make([]nexusiq.License, 0, len(licensesSet))
	for l := range licensesSet {
		switch l.LicenseID {
		case "Not-Supported", "No-Source-License", "Not-Declared", "No-Sources", "See-License-Clause", "UNKNOWN":
			continue
		}
		licenses = append(licenses, l)
	}

	switch {
	// Serve the HTML page on the given port
	case *servePortPtr != "":
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if err := writeHTML(w, *appIDPtr, licenses); err != nil {
				log.Printf("could not serve page: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		})

		log.Printf("Serving licenses report on port %s\n", *servePortPtr)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *servePortPtr), nil))
	// Create the HTML page as a file
	case *filePtr != "":
		fo, err := os.Create(*filePtr)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		w := bufio.NewWriter(fo)

		if err := writeHTML(w, *appIDPtr, licenses); err != nil {
			panic(err)
		}

		if err = w.Flush(); err != nil {
			panic(err)
		}
	default:
		panic("Missing required argument. One of 'file' or 'serve' must be used. See -help")
	}
}
