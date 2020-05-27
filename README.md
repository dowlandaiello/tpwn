# tpwn
TestPwn - A command-line utility for scraping online school test questions from quizlet.com.

## What is this?

Tpwn is a simple command-line utility that takes a question and searches quizlet.com for the answer. To get started, install the binary with `go install`:

```zsh
go install github.com/dowlandaiello/tpwn
```

## How do I use it?

Simple: `tpwn your_question`. For example:

```zsh
tpwn "Which of the following represents an example of a population?"
```

or, if you'd like to read questions from a file:

```zsh
tpwn file_name.txt
```

Results will be written to stdout, which can be used with standard unix operations like piping and writing to a file:

```zsh
tpwn "Which of the following represents an example of a population?" | xclip
tpwn questions.txt > answers.txt
```
