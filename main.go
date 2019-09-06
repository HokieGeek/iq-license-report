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

	nexusiq "github.com/sonatype-nexus-community/gonexus/iq"
)

var pageTmpl = `<html>
	<head>
		<title>{{ .Title }}</title>
	</head>
	<body>
		<ul>
		{{ range .Licenses }}
			<li>({{ .LicenseID }}) {{ .LicenseName }}</li>
		{{ end }}
		</ul>
	</body>
</html>`

func writeHTML(w io.Writer, appID string, licenses []nexusiq.License) error {
	page := struct {
		Title    string
		Licenses []nexusiq.License
	}{
		Title:    fmt.Sprintf("Licenses for %s", appID),
		Licenses: licenses,
	}

	t := template.Must(template.New("page").Parse(pageTmpl))
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

	auth := strings.Split(*iqAuthPtr, ":")
	iq, err := nexusiq.New(*iqServerPtr, auth[0], auth[1])
	if err != nil {
		panic(err)
	}

	report, err := nexusiq.GetRawReportByAppID(iq, *appIDPtr, *stagePtr)
	if err != nil {
		panic(err)
	}

	licensesSet := make(map[nexusiq.License]struct{})
	for _, c := range report.Components {
		for _, l := range c.LicensesData.DeclaredLicenses {
			licensesSet[l] = struct{}{}
		}
		for _, l := range c.LicensesData.ObservedLicenses {
			licensesSet[l] = struct{}{}
		}
	}

	licenses := make([]nexusiq.License, 0, len(licensesSet))
	for l := range licensesSet {
		if l.LicenseID == "Not-Supported" {
			continue
		}
		licenses = append(licenses, l)
	}

	switch {
	case *servePortPtr != "":
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if err := writeHTML(w, *appIDPtr, licenses); err != nil {
				log.Printf("could not serve page: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		})

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *servePortPtr), nil))
	case *filePtr != "":
		fmt.Println(report.Components[0].LicensesData)

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

		writeHTML(w, *appIDPtr, nil)
	default:
		panic("One of file or serve must be used. See help")
	}
}
