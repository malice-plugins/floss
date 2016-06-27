package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"
	"github.com/urfave/cli"
	r "gopkg.in/dancannon/gorethink.v2"
)

// Version stores the plugin's version
var Version string

// BuildTime stores the plugin's build time
var BuildTime string

const (
	name     = "floss"
	category = "pe"
)

type pluginResults struct {
	ID   string      `json:"id" gorethink:"id,omitempty"`
	Data resultsData `json:"floss" gorethink:"floss"`
}

type floss struct {
	Results resultsData `json:"floss"`
}

type resultsData struct {
	DecodedStrings []decodedStrings `json:"decoded" gorethink:"decoded"`
	StackStrings   []string         `json:"stack" gorethink:"stack"`
}

type decodedStrings struct {
	Location string   `json:"location" gorethink:"location"`
	Strings  []string `json:"strings" gorethink:"strings"`
}

func getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}

func assert(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// getSHA256 calculates a file's sha256sum
func getSHA256(name string) string {

	dat, err := ioutil.ReadFile(name)
	assert(err)

	h256 := sha256.New()
	_, err = h256.Write(dat)
	assert(err)

	return fmt.Sprintf("%x", h256.Sum(nil))
}

// RunCommand runs cmd on file
func RunCommand(cmd string, args ...string) string {

	cmdOut, err := exec.Command(cmd, args...).Output()
	if len(cmdOut) == 0 {
		assert(err)
	}

	return string(cmdOut)
}

func printStatus(resp gorequest.Response, body string, errs []error) {
	fmt.Println(resp.Status)
}

func printMarkDownTable(f floss) {
	fmt.Println("#### Floss")
	fmt.Println("##### Decoded Strings")
	if f.Results.DecodedStrings != nil {
		for _, decodedStr := range f.Results.DecodedStrings {
			fmt.Printf("Location: `%s`\n", decodedStr.Location)
			for _, dStr := range decodedStr.Strings {
				fmt.Printf(" - `%s`\n", dStr)
			}
		}
	} else {
		fmt.Println(" - No Strings")
	}
	fmt.Println("##### Stack Strings")
	if f.Results.StackStrings != nil {
		for _, stkStr := range f.Results.StackStrings {
			fmt.Printf(" - `%s`\n", stkStr)
		}
	} else {
		fmt.Println(" - No Strings")
	}
}

func removeDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func getLocationAndNumOfDecodedStrs(line string) (string, int) {
	line = strings.TrimPrefix(line, "Decoding function at")
	location := strings.TrimSpace(strings.Split(line, "(decoded")[0])
	line = strings.Split(line, "(decoded")[1]
	num, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(line, "strings)")))
	assert(err)

	return location, num
}

func parseFlossOutput(flossOutput string) resultsData {

	keepLines := []string{}
	results := resultsData{}
	var decodedStrArray []decodedStrings

	lines := strings.Split(flossOutput, "\n")
	// remove empty lines
	for _, line := range lines {
		if len(strings.TrimSpace(line)) != 0 {
			keepLines = append(keepLines, strings.TrimSpace(line))
		}
	}
	// build results data
	for i := 0; i < len(keepLines); i++ {
		if strings.Contains(keepLines[i], "Decoding function at") {
			// get function location
			location, numOfStrings := getLocationAndNumOfDecodedStrs(keepLines[i])

			decodedStr := decodedStrings{
				Location: location,
				Strings:  removeDuplicates(keepLines[i+1 : i+numOfStrings+1]),
			}
			decodedStrArray = append(decodedStrArray, decodedStr)
			i = i + numOfStrings
			continue
		} else if strings.Contains(keepLines[i], "stackstrings") {
			// get stackstrings
			line := strings.TrimPrefix(keepLines[i], "FLOSS extracted")
			numOfStrings, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(line, "stackstrings")))
			assert(err)
			// if len(keepLines) < i+numOfStrings+1 {
			results.StackStrings = keepLines[i+1 : i+numOfStrings+1]
			// } else {
			// 	log.Fatal("shiz went sideways.")
			// }
			i = i + numOfStrings
			continue
		}
	}
	results.DecodedStrings = decodedStrArray

	return results
}

// scanFile scans file with all floss rules in the rules folder
func scanFile(path string) floss {
	flossResults := floss{}
	flossResults.Results = parseFlossOutput(RunCommand("/usr/bin/floss", "-g", path))

	return flossResults
}

// writeToDatabase upserts plugin results into Database
func writeToDatabase(results pluginResults) {

	// connect to RethinkDB
	session, err := r.Connect(r.ConnectOpts{
		Address:  fmt.Sprintf("%s:28015", getopt("MALICE_RETHINKDB", "rethink")),
		Timeout:  5 * time.Second,
		Database: "malice",
	})
	if err != nil {
		log.Debug(err)
		return
	}
	defer session.Close()

	res, err := r.Table("samples").Get(results.ID).Run(session)
	assert(err)
	defer res.Close()

	if res.IsNil() {
		// upsert into RethinkDB
		resp, err := r.Table("samples").Insert(results, r.InsertOpts{Conflict: "replace"}).RunWrite(session)
		assert(err)
		log.Debug(resp)
	} else {
		resp, err := r.Table("samples").Get(results.ID).Update(map[string]interface{}{
			"plugins": map[string]interface{}{
				category: map[string]interface{}{
					name: results.Data,
				},
			},
		}).RunWrite(session)
		assert(err)

		log.Debug(resp)
	}
}

var appHelpTemplate = `Usage: {{.Name}} {{if .Flags}}[OPTIONS] {{end}}COMMAND [arg...]

{{.Usage}}

Version: {{.Version}}{{if or .Author .Email}}

Author:{{if .Author}}
  {{.Author}}{{if .Email}} - <{{.Email}}>{{end}}{{else}}
  {{.Email}}{{end}}{{end}}
{{if .Flags}}
Options:
  {{range .Flags}}{{.}}
  {{end}}{{end}}
Commands:
  {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
  {{end}}
Run '{{.Name}} COMMAND --help' for more information on a command.
`

func init() {
	log.SetLevel(log.ErrorLevel)
}

func main() {
	cli.AppHelpTemplate = appHelpTemplate
	app := cli.NewApp()
	app.Name = "floss"
	app.Author = "blacktop"
	app.Email = "https://github.com/blacktop"
	app.Version = Version + ", BuildTime: " + BuildTime
	app.Compiled, _ = time.Parse("20060102", BuildTime)
	app.Usage = "Malice FLOSS Plugin"
	var table bool
	var rethinkdb string
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "verbose output",
		},
		cli.StringFlag{
			Name:        "rethinkdb",
			Value:       "",
			Usage:       "rethinkdb address for Malice to store results",
			EnvVar:      "MALICE_RETHINKDB",
			Destination: &rethinkdb,
		},
		cli.BoolFlag{
			Name:   "post, p",
			Usage:  "POST results to Malice webhook",
			EnvVar: "MALICE_ENDPOINT",
		},
		cli.BoolFlag{
			Name:   "proxy, x",
			Usage:  "proxy settings for Malice webhook endpoint",
			EnvVar: "MALICE_PROXY",
		},
		cli.BoolFlag{
			Name:        "table, t",
			Usage:       "output as Markdown table",
			Destination: &table,
		},
	}
	app.ArgsUsage = "FILE to scan with FLOSS"
	app.Action = func(c *cli.Context) error {
		if c.Args().Present() {
			path := c.Args().First()
			// Check that file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				assert(err)
			}

			if c.Bool("verbose") {
				log.SetLevel(log.DebugLevel)
			}

			floss := scanFile(path)

			// upsert into Database
			writeToDatabase(pluginResults{
				ID:   getopt("MALICE_SCANID", getSHA256(path)),
				Data: floss.Results,
			})

			if table {
				printMarkDownTable(floss)
			} else {
				flossJSON, err := json.Marshal(floss)
				assert(err)
				fmt.Println(string(flossJSON))
			}
		} else {
			log.Fatal(fmt.Errorf("Please supply a file to scan with FLOSS"))
		}
		return nil
	}

	err := app.Run(os.Args)
	assert(err)
}
