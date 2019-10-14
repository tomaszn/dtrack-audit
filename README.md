# dtrack-audit
[OWASP Dependency Track](https://dependencytrack.org) API client. See [Dependency-Track docs: Continuous Integration & Delivery](https://docs.dependencytrack.org/usage/cicd/) for use case.

## Install

### Local Installation

```bash
go get github.com/ozonru/dtrack-audit/cmd/dtrack-audit
```

## Usage

```bash
$ dtrack-audit -h

Send SBOM file to Dependency Track.

Usage of program:
  -i string
        Target SBOM file (default "bom.xml")
  -k string
        API Key (Required) (default is environment variable $DTRACK_API_KEY)
  -p string
        Project ID (Required) (default is environment variable $DTRACK_PROJECT_ID)
  -s    Sync mode enabled (Upload SBOM file and wait for scan result)
  -t int
        Max timeout in second for polling API for project findings (default 25)
  -u string
        API URL (Required) (default is environment variable $DTRACK_API_URL)
```

### Sample output

```bash
$ cyclonedx-bom -o bom.xml
$ dtrack-audit -s

2019/10/14 13:04:34 SBOM file is successfully uploaded to DTrack API. Result token is 12345f5e-4ccb-45fe-b8fd-1234a8bf0081

2 vulnerabilities found!

 > HIGH: Arbitrary File Write
   Component: adm-zip 0.4.7
   More info: https://dtrack/vulnerability/?source=NPM&vulnId=994

 > CRITICAL: Prototype Pollution
   Component: handlebars 4.0.11
   More info: https://dtrack/vulnerability/?source=NPM&vulnId=755
```