package main

import (
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/urfave/cli/v2"

	"bufio"
	"fmt"
	"log"
	"os"
)

// MaxSources is the maximum number of quizlet sets that will be scanned.
const MaxSources = 5

// QueryType represents all possible types of queries.
type QueryType int

const (
	// CommandLineArguments represents a query sourced from a list of command-line arguments.
	CommandLineArguments QueryType = iota

	// File represents a query sourced from a file.
	File
)

// Query is an iterator representing questions submitted by the user.
type Query struct {
	queryType QueryType // the type of query

	// the source of the query arguments
	args    cli.Args
	scanner *bufio.Scanner

	index int // the position of the iterator
}

// NextQuestion gets the next question stored in the query.
func (q *Query) NextQuestion() (question string) {
	switch q.queryType {
	case CommandLineArguments:
		question = q.args.Get(q.index)

		q.index++

		return
	case File:
		q.scanner.Scan()
		question = q.scanner.Text()
	}

	return
}

// QueryFromContext parses a urfave/cli context instance into a query.
func QueryFromContext(c *cli.Context) (q *Query) {
	args := c.Args()

	q = &Query{
		queryType: 0,
		args:      nil,
		scanner:   nil,
		index:     0,
	}

	if f, err := os.Open(args.Get(0)); err == nil {
		args = nil

		q.queryType = File
		q.scanner = bufio.NewScanner(f)

		return
	}

	q.queryType = CommandLineArguments
	q.args = args

	return
}

func main() {
	app := &cli.App{
		Name:  "tpwn",
		Usage: "parse a dump of numbered questions, and search for appropriate answers on quizlet.com",
		Action: func(c *cli.Context) error {
			// Load questions from CLI args or from a local file
			q := QueryFromContext(c)
			question := q.NextQuestion()

			for len(question) > 0 {
				answer, err := GetAnswer(question)
				if err != nil {
					log.Fatal(err)
				}

				fmt.Println(answer)

				question = q.NextQuestion()
			}

			return nil
		},
	}

	app.Run(os.Args)
}

// GetAnswer searches quizlet for the provided question.
func GetAnswer(question string) (answer string, err error) {
	googleURL := fmt.Sprintf("https://www.google.com/search?client=firefox-b-1-d&q=%s", strings.Replace(question, " ", "+", -1))

	occurrences := make(map[string]int)

	// Don't recurse, since that wouldn't bring us to the right study set
	c := colly.NewCollector(colly.MaxDepth(0), colly.Async(true))
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 5})

	i := 0

	// Load a bunch of quizlet study sets from google
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.Contains(link, "quizlet.com") && !strings.Contains(link, "create-set") && i < MaxSources {
			c.OnRequest(func(r *colly.Request) {
				r.Headers.Set("Accept", "*/*")
				r.Headers.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.122 Safari/537.36")
			})

			lParts := strings.Split(link, "=")
			if len(lParts) > 1 {
				link = lParts[1]
			}

			i++

			c.Visit(link)
		}
	})

	// Look for the answer to the question
	c.OnHTML(".SetPageTerm-inner", func(e *colly.HTMLElement) {
		// Make sure the question appears somewhere on the page, look for where it appears
		if strings.Contains(strings.ToLower(e.ChildText(".SetPageTerm-wordText")), strings.ToLower(question)) {
			// Get the answer to the question by its class
			possibleAnswer := e.ChildText(".SetPageTerm-definitionText")
			occurrences[possibleAnswer]++

			// For each worker that's using a different study set, make ensure the given answer is the most popular
			if occurrences[possibleAnswer] > occurrences[answer] {
				answer = possibleAnswer
			}
		}
	})

	// Write any errors to the output buffer
	c.OnError(func(_ *colly.Response, reqErr error) {
		err = reqErr
	})

	c.Visit(googleURL)
	c.Wait()

	return
}
