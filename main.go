package main

import (
	"github.com/gocolly/colly/v2"
	"github.com/urfave/cli/v2"

	"bufio"
	"fmt"
	"os"
	"strings"
)

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
			q := QueryFromContext(c)
			question := q.NextQuestion()

			for len(question) > 0 {
				fmt.Println(GetAnswer(question))

				question = q.NextQuestion()
			}

			return nil
		},
	}

	app.Run(os.Args)
}

// GetAnswer searches quizlet for the provided question.
func GetAnswer(question string) (string, error) {
	googleURL := fmt.Sprintf("https://www.google.com/search?client=firefox-b-1-d&q=%s", strings.Replace(question, " ", "+", -1))

	c := colly.NewCollector(colly.MaxDepth(2))

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		if strings.Contains(link, "quizlet.com") {
			fmt.Println(link)

			c.Visit(strings.Split(link, "=")[1])
		}
	})

	c.OnHTML(".SetPageTerm-definitionText", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
	})

	c.Visit(googleURL)

	return "", nil
}
