# iq-license-report

## Building

You can either use the binaries that are - or should be - associated with the latest release on this GH page, or you can run `go build .`

## Usage

```sh
./iq-license-report -appId <publicAppID>  [-iq http://<iqserver>:<iqport>] [-auth username:password] [-stage <iqstage>] [-file <filename>] [-serve <port>]
```

One of `-file` or `-serve` is required

See the IQ documentation to find the application public ID from the GUI:
<https://help.sonatype.com/iqserver/managing/application-management#ApplicationManagement-CopyingtheApplicationIDtoClipboard>

Example of creating an HTML report as a file:

```sh
./iq-license-report -iq http://localhost:8070 -auth admin:admin123 -appId awesomeApp -stage release -file awesomeApp-licenses.html
```
