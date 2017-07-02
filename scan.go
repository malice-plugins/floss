package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/fatih/structs"
	"github.com/gorilla/mux"
	"github.com/malice-plugins/go-plugin-utils/database/elasticsearch"
	"github.com/malice-plugins/go-plugin-utils/utils"
	"github.com/parnurzeal/gorequest"
	"github.com/urfave/cli"
)

// Version stores the plugin's version
var Version string

// BuildTime stores the plugin's build time
var BuildTime string

var path string

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
	MarkDown       string           `json:"markdown" structs:"markdown"`
}

type decodedStrings struct {
	Location string   `json:"location" structs:"location"`
	Strings  []string `json:"strings" structs:"strings"`
}

func assert(err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"plugin":   name,
			"category": category,
			"path":     path,
		}).Fatal(err)
	}
}

// scanFile scans file with all floss rules in the rules folder
func scanFile(timeout int, all bool) floss {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	flossResults := floss{}
	// flossResults.Results = parseFlossOutput(RunCommand("./floss", "-g", path), all)
	output, err := utils.RunCommand(ctx, "/usr/bin/floss", "-g", path)
	assert(err)
	flossResults.Results = parseFlossOutput(output, err, all)

	return flossResults
}

func parseFlossOutput(flossOutput string, err error, all bool) resultsData {

	if err != nil {
		return resultsData{ASCIIStrings: []string{err.Error()}}
	}

	log.WithFields(log.Fields{
		"plugin":   name,
		"category": category,
		"path":     path,
	}).Debug("FLOSS Output: ", flossOutput)

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
				results.ASCIIStrings = utils.RemoveDuplicates(getASCIIStrings(keepLines[i+1 : len(keepLines)]))
			}
			if strings.Contains(keepLines[i], "FLOSS static UTF-16 strings") {
				results.UTF16Strings = utils.RemoveDuplicates(getUTF16Strings(keepLines[i+1 : len(keepLines)]))
			}
		}
		if strings.Contains(keepLines[i], "Decoding function at") {
			// get function location
			location, numOfStrings := getLocationAndNumOfDecodedStrs(keepLines[i])

			decodedStr := decodedStrings{
				Location: location,
				Strings:  utils.RemoveDuplicates(keepLines[i+1 : i+numOfStrings+1]),
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

func getLocationAndNumOfDecodedStrs(line string) (string, int) {
	numMatch := regexp.MustCompile("[0-9]+").FindAllString(line, -1)
	locMatch := regexp.MustCompile("0x[0-9A-F]+").FindAllString(line, -1)

	if len(locMatch) > 0 && len(numMatch) > 0 {
		num, err := strconv.Atoi(numMatch[len(numMatch)-1])
		assert(err)

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

func generateMarkDownTable(f floss) string {
	var tplOut bytes.Buffer

	t := template.Must(template.New("floss").Parse(tpl))

	err := t.Execute(&tplOut, f)
	if err != nil {
		log.Println("executing template:", err)
	}

	return tplOut.String()

	// fmt.Printf("#### Floss\n\n")
	// if f.Results.ASCIIStrings != nil {
	// 	fmt.Printf("##### ASCII Strings\n\n")
	// 	for _, ascStr := range f.Results.ASCIIStrings {
	// 		fmt.Printf(" - `%s`\n", ascStr)
	// 	}
	// 	fmt.Println()
	// }
	// if f.Results.UTF16Strings != nil {
	// 	fmt.Printf("##### UTF-16 Strings\n\n")
	// 	for _, utfStr := range f.Results.UTF16Strings {
	// 		fmt.Printf(" - `%s`\n", utfStr)
	// 	}
	// 	fmt.Println()
	// }
	// fmt.Printf("##### Decoded Strings\n\n")
	// if f.Results.DecodedStrings != nil {
	// 	for _, decodedStr := range f.Results.DecodedStrings {
	// 		fmt.Printf("Location: `%s`\n", decodedStr.Location)
	// 		for _, dStr := range decodedStr.Strings {
	// 			fmt.Printf(" - `%s`\n", dStr)
	// 		}
	// 		fmt.Println()
	// 	}
	// } else {
	// 	fmt.Println(" - No Strings")
	// }
	// fmt.Printf("##### Stack Strings\n\n")
	// if f.Results.StackStrings != nil {
	// 	for _, stkStr := range f.Results.StackStrings {
	// 		fmt.Printf(" - `%s`\n", stkStr)
	// 	}
	// } else {
	// 	fmt.Println(" - No Strings")
	// }
}

func printMarkDownTable(f floss) {
	fmt.Printf(generateMarkDownTable(f))
}

func printStatus(resp gorequest.Response, body string, errs []error) {
	fmt.Println(body)
}

func webService() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/scan", webAvScan).Methods("POST")
	log.Info("web service listening on port :3993")
	log.Fatal(http.ListenAndServe(":3993", router))
}

func webAvScan(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	all := vars["all"]

	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("malware")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "Please supply a valid file to scan.")
		log.Error(err)
	}
	defer file.Close()

	log.Debug("Uploaded fileName: ", header.Filename)

	tmpfile, err := ioutil.TempFile("/malware", "web_")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	data, err := ioutil.ReadAll(file)

	if _, err = tmpfile.Write(data); err != nil {
		log.Fatal(err)
	}
	if err = tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	// Do AV scan
	var fl floss
	path = tmpfile.Name()
	if strings.EqualFold(all, "true") || all != "" {
		fl = scanFile(60, true)
	} else {
		fl = scanFile(60, false)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(fl); err != nil {
		log.Fatal(err)
	}
}

func main() {

	var elastic string

	cli.AppHelpTemplate = utils.AppHelpTemplate
	app := cli.NewApp()

	app.Name = "floss"
	app.Author = "blacktop"
	app.Email = "https://github.com/blacktop"
	app.Version = Version + ", BuildTime: " + BuildTime
	app.Compiled, _ = time.Parse("20060102", BuildTime)
	app.Usage = "Malice FLOSS Plugin"
	app.ArgsUsage = "FILE to scan with FLOSS"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "verbose output",
		},
		cli.IntFlag{
			Name:   "timeout",
			Value:  120,
			Usage:  "malice plugin timeout (in seconds)",
			EnvVar: "MALICE_TIMEOUT",
		},
		cli.StringFlag{
			Name:        "elasitcsearch",
			Value:       "",
			Usage:       "elasitcsearch address for Malice to store results",
			EnvVar:      "MALICE_ELASTICSEARCH",
			Destination: &elastic,
		},
		cli.BoolFlag{
			Name:   "callback, c",
			Usage:  "POST results to Malice webhook",
			EnvVar: "MALICE_ENDPOINT",
		},
		cli.BoolFlag{
			Name:   "proxy, x",
			Usage:  "proxy settings for Malice webhook endpoint",
			EnvVar: "MALICE_PROXY",
		},
		cli.BoolFlag{
			Name:  "table, t",
			Usage: "output as Markdown table",
		},
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "output ascii/utf-16 strings",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "web",
			Usage: "Create a FLOSS scan web service",
			Action: func(c *cli.Context) error {
				webService()
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {

		var err error

		if c.Bool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		if c.Args().Present() {
			path, err = filepath.Abs(c.Args().First())
			assert(err)

			if _, err = os.Stat(path); os.IsNotExist(err) {
				assert(err)
			}

			floss := scanFile(c.Int("timeout"), c.Bool("all"))

			// upsert into Database
			elasticsearch.InitElasticSearch(elastic)
			elasticsearch.WritePluginResultsToDatabase(elasticsearch.PluginResults{
				ID:       utils.Getopt("MALICE_SCANID", utils.GetSHA256(path)),
				Name:     name,
				Category: category,
				Data:     structs.Map(floss.Results),
			})

			if c.Bool("table") {
				printMarkDownTable(floss)
			} else {
				flossJSON, err := json.Marshal(floss)
				assert(err)
				if c.Bool("callback") {
					request := gorequest.New()
					if c.Bool("proxy") {
						request = gorequest.New().Proxy(os.Getenv("MALICE_PROXY"))
					}
					request.Post(os.Getenv("MALICE_ENDPOINT")).
						Set("X-Malice-ID", utils.Getopt("MALICE_SCANID", utils.GetSHA256(path))).
						Send(string(flossJSON)).
						End(printStatus)

					return nil
				}
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
