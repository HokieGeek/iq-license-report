package iqlicensereport

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	nexusiq "github.com/sonatype-nexus-community/gonexus/iq"
)

var pageTmpl = `<html>
	<head>
		<title>{{ .Title }}</title>
	</head>
	<body>
		<ul>
		{{ range .Licenses }}
			<li>{{ .LicenseName }}</li>
		{{ end }}
		</ul>
	</body>
</html>`

// WriteHTML writes the reports in the database to an io.Writer after applying an HTML template
func WriteHTML(w io.Writer, licenses []nexusiq.License) error {
	page := struct {
		Title    string
		Licenses []nexusiq.License
	}{
		Title:    "foo",
		Licenses: licenses,
	}

	t := template.Must(template.New("page").Parse(pageTmpl))
	if err := t.Execute(w, page); err != nil {
		return fmt.Errorf("could not generate page: %v", err)
	}

	return nil
}

// HTMLLicensesList serves a page with all of the reports in the database
func HTMLLicensesList(w http.ResponseWriter, r *http.Request, licenses []nexusiq.License) {
	if err := WriteHTML(w, licenses); err != nil {
		log.Printf("could not serve page: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
