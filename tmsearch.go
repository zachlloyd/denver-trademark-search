package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/urfave/cli"
	"golang.org/x/net/html"
)

func main() {
	var tmsearchURL string
	var numResults int

	app := &cli.App{
		Name:  "tmsearch",
		Usage: "Scrape records from the tmsearch database for particular search terms",
		Action: func(c *cli.Context) error {
			scrape(tmsearchURL, numResults)
			return nil
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "tmsearchURL",
				Value:       "<unset>",
				Usage:       "The root url for scraping results",
				Destination: &tmsearchURL,
			},
			&cli.IntFlag{
				Name:        "numResults",
				Value:       1,
				Usage:       "The root url for scraping results",
				Destination: &numResults,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Results from a freeform search on http://tmsearch.uspto.gov/ for
// (neon$)[FM]  and  (live)[LD] and (009)[CC]

func scrape(tmsearchURL string, numResults int) {
	client := &http.Client{
		CheckRedirect: http.DefaultClient.CheckRedirect,
	}
	csvWriter := csv.NewWriter(os.Stdout)

	for i := 1; i <= numResults; i++ {
		url := tmsearchURL + strconv.Itoa(i)
		req := newRequest(url)
		log.Println("Scraping results from url", url)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Error reading response", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			bodyString := string(bodyBytes)
			processResult(i, csvWriter, bodyString)
		}
		time.Sleep(2 * time.Second)
	}
}

func processResult(resultNum int, csvWriter *csv.Writer, body string) {
	doc, err := htmlquery.Parse(strings.NewReader(body))
	if err != nil {
		log.Fatal("Unable to parse html doc", err)
		return
	}

	writeLine(csvWriter, doc, htmlquery.Find(doc, "//table[4]/tbody/tr/td[1]/b"))
	writeLine(csvWriter, doc, htmlquery.Find(doc, "//table[4]/tbody/tr/td[2]"))
	csvWriter.Flush()
}

func writeLine(csvWriter *csv.Writer, doc *html.Node, nodes []*html.Node) {
	cols := make([]string, 0)
	for _, node := range nodes {
		cols = append(cols, htmlquery.InnerText(node))
	}
	csvWriter.Write(cols)
}

func newRequest(url string) *http.Request {
	content, err := ioutil.ReadFile("cookie.txt")
	if err != nil {
		log.Fatal("Error reading cookie file", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	cookies := strings.Split(string(content), ";")
	for _, cookie := range cookies {
		keyValue := strings.Split(cookie, "=")
		req.AddCookie(&http.Cookie{
			Name:  strings.TrimSpace(keyValue[0]),
			Value: strings.TrimSpace(keyValue[1]),
		})
	}
	return req
}
