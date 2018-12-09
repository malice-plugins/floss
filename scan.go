package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
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
	"github.com/malice-plugins/pkgs/database"
	"github.com/malice-plugins/pkgs/database/elasticsearch"
	"github.com/malice-plugins/pkgs/utils"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	name     = "floss"
	category = "exe"
)

var (
	// Version stores the plugin's version
	Version string
	// BuildTime stores the plugin's build time
	BuildTime string

	path    string
	timeout int
	// es is the elasticsearch database object
	es elasticsearch.Database
)

type pluginResults struct {
	ID   string      `json:"id" structs:"id,omitempty"`
	Data resultsData `json:"floss" structs:"floss"`
}

type pluginError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type floss struct {
	Results resultsData `json:"floss,omitempty"`
	Error   pluginError `json:"error,omitempty"`
}

type resultsData struct {
	ASCIIStrings   []string         `json:"ascii,omitempty" structs:"ascii"`
	UTF16Strings   []string         `json:"utf-16,omitempty" structs:"utf-16"`
	DecodedStrings []decodedStrings `json:"decoded,omitempty" structs:"decoded"`
	StackStrings   []string         `json:"stack,omitempty" structs:"stack"`
	MarkDown       string           `json:"markdown,omitempty" structs:"markdown,omitempty"`
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
func scanFile(ctx context.Context, all bool) floss {

	flossResults := floss{}

	output, err := exec.CommandContext(ctx, "/usr/bin/floss", "-g", path).Output()
	if ctx.Err() == context.DeadlineExceeded {
		return floss{Error: pluginError{
			Message: errors.Wrap(ctx.Err(), "plugin deadline exceeded (increase timeout with --timeout flag)").Error(),
			Type:    "timeout",
		}}
	}
	if err != nil {
		return floss{Error: pluginError{
			Message: errors.Wrapf(err, "cmd failed: /usr/bin/floss -g %s", path).Error(),
			Type:    "exec",
		}}
	}

	flossResults.Results = parseFlossOutput(string(output), all)

	return flossResults
}

func parseFlossOutput(flossOutput string, all bool) resultsData {

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

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

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
		fl = scanFile(ctx, true)
	} else {
		fl = scanFile(ctx, false)
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if fl.Error != (pluginError{}) {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(fl); err != nil {
		log.Fatal(err)
	}
}

func main() {

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
			Name:        "timeout",
			Value:       240,
			Usage:       "malice plugin timeout (in seconds)",
			EnvVar:      "MALICE_TIMEOUT",
			Destination: &timeout,
		},
		cli.StringFlag{
			Name:        "elasticsearch",
			Value:       "",
			Usage:       "elasticsearch url for Malice to store results",
			EnvVar:      "MALICE_ELASTICSEARCH_URL",
			Destination: &es.URL,
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

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		csig := make(chan os.Signal, 1)
		signal.Notify(csig, os.Interrupt)
		defer func() {
			signal.Stop(csig)
			cancel()
		}()
		go func() {
			select {
			case <-csig:
				cancel()
			case <-ctx.Done():
			}
		}()

		if c.Bool("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		if c.Args().Present() {
			path, err = filepath.Abs(c.Args().First())
			assert(err)

			if _, err = os.Stat(path); os.IsNotExist(err) {
				assert(err)
			}

			floss := scanFile(ctx, c.Bool("all"))
			// log.Debug(floss)
			// if floss.Error != (pluginError{}){
			// 	return errors.Wrapf(floss.Error.Error, "failed to scan %s", path)
			// }
			floss.Results.MarkDown = generateMarkDownTable(floss)

			// upsert into Database
			if len(c.String("elasticsearch")) > 0 {
				err := es.Init()
				if err != nil {
					return errors.Wrap(err, "failed to initalize elasticsearch")
				}
				err = es.StorePluginResults(database.PluginResults{
					ID:       utils.Getopt("MALICE_SCANID", utils.GetSHA256(path)),
					Name:     name,
					Category: category,
					Data:     structs.Map(floss.Results),
				})
				if err != nil {
					return errors.Wrapf(err, "failed to index malice/%s results", name)
				}
			}

			if c.Bool("table") {
				fmt.Printf(floss.Results.MarkDown)
			} else {
				floss.Results.MarkDown = ""
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
