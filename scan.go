package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"github.com/maliceio/go-plugin-utils/database/elasticsearch"
	"github.com/maliceio/go-plugin-utils/utils"
	"github.com/urfave/cli"
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
	ID   string      `json:"id" structs:"id,omitempty"`
	Data resultsData `json:"floss" structs:"floss"`
}

type floss struct {
	Results resultsData `json:"floss"`
}

type resultsData struct {
	ASCIIStrings   []string         `json:"ascii" structs:"ascii"`
	UTF16Strings   []string         `json:"utf-16" structs:"utf-16"`
	DecodedStrings []decodedStrings `json:"decoded" structs:"decoded"`
	StackStrings   []string         `json:"stack" structs:"stack"`
}

type decodedStrings struct {
	Location string   `json:"location" structs:"location"`
	Strings  []string `json:"strings" structs:"strings"`
}

func printMarkDownTable(f floss) {
	fmt.Printf("#### Floss\n\n")
	if f.Results.ASCIIStrings != nil {
		fmt.Printf("##### ASCII Strings\n\n")
		for _, ascStr := range f.Results.ASCIIStrings {
			fmt.Printf(" - `%s`\n", ascStr)
		}
		fmt.Println()
	}
	if f.Results.UTF16Strings != nil {
		fmt.Printf("##### UTF-16 Strings\n\n")
		for _, utfStr := range f.Results.UTF16Strings {
			fmt.Printf(" - `%s`\n", utfStr)
		}
		fmt.Println()
	}
	fmt.Printf("##### Decoded Strings\n\n")
	if f.Results.DecodedStrings != nil {
		for _, decodedStr := range f.Results.DecodedStrings {
			fmt.Printf("Location: `%s`\n", decodedStr.Location)
			for _, dStr := range decodedStr.Strings {
				fmt.Printf(" - `%s`\n", dStr)
			}
			fmt.Println()
		}
	} else {
		fmt.Println(" - No Strings")
	}
	fmt.Printf("##### Stack Strings\n\n")
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
	numMatch := regexp.MustCompile("[0-9]+").FindAllString(line, -1)
	locMatch := regexp.MustCompile("0x[0-9A-F]+").FindAllString(line, -1)

	if len(locMatch) > 0 && len(numMatch) > 0 {
		num, err := strconv.Atoi(numMatch[len(numMatch)-1])
		utils.Assert(err)

		return locMatch[0], num
	}
	return "", 0
}

func getASCIIStrings(strArray []string) []string {
	asciiStrings := []string{}
	for _, str := range strArray {
		if strings.Contains(str, "FLOSS static UTF-16 strings") {
			break
		}
		asciiStrings = append(asciiStrings, str)
	}
	return asciiStrings
}

func getUTF16Strings(strArray []string) []string {
	utf16Strings := []string{}
	for _, str := range strArray {
		if strings.Contains(str, "FLOSS decoded") {
			break
		}
		utf16Strings = append(utf16Strings, str)
	}
	return utf16Strings
}

func parseFlossOutput(flossOutput string, all bool) resultsData {

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
		if all {
			if strings.Contains(keepLines[i], "FLOSS static ASCII strings") {
				results.ASCIIStrings = removeDuplicates(getASCIIStrings(keepLines[i+1 : len(keepLines)]))
			}
			if strings.Contains(keepLines[i], "FLOSS static UTF-16 strings") {
				results.UTF16Strings = removeDuplicates(getUTF16Strings(keepLines[i+1 : len(keepLines)]))
			}
		}
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
			utils.Assert(err)
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
func scanFile(path string, all bool) floss {
	flossResults := floss{}
	// flossResults.Results = parseFlossOutput(RunCommand("./floss", "-g", path), all)
	flossResults.Results = parseFlossOutput(utils.RunCommand("/usr/bin/floss", "-g", path), all)

	return flossResults
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
	var all bool
	var elasitcsearch string
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "verbose output",
		},
		cli.StringFlag{
			Name:        "elasitcsearch",
			Value:       "",
			Usage:       "elasitcsearch address for Malice to store results",
			EnvVar:      "MALICE_ELASTICSEARCH",
			Destination: &elasitcsearch,
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
		cli.BoolFlag{
			Name:        "all, a",
			Usage:       "output ascii/utf-16 strings",
			Destination: &all,
		},
	}
	app.ArgsUsage = "FILE to scan with FLOSS"
	app.Action = func(c *cli.Context) error {
		if c.Args().Present() {
			path := c.Args().First()
			// Check that file exists
			if _, err := os.Stat(path); os.IsNotExist(err) {
				utils.Assert(err)
			}

			if c.Bool("verbose") {
				log.SetLevel(log.DebugLevel)
			}

			floss := scanFile(path, all)

			// upsert into Database
			elasticsearch.InitElasticSearch()
			elasticsearch.WritePluginResultsToDatabase(elasticsearch.PluginResults{
				ID:       utils.Getopt("MALICE_SCANID", utils.GetSHA256(path)),
				Name:     name,
				Category: category,
				Data:     structs.Map(floss.Results),
			})

			if table {
				printMarkDownTable(floss)
			} else {
				flossJSON, err := json.Marshal(floss)
				utils.Assert(err)
				fmt.Println(string(flossJSON))
			}
		} else {
			log.Fatal(fmt.Errorf("Please supply a file to scan with FLOSS"))
		}
		return nil
	}

	err := app.Run(os.Args)
	utils.Assert(err)
}
