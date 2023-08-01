package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// TODO(Nick):
// 1) Do I even need the markdown parser?
// 2) Unmarshalling the json file is almost as slow as parsing and creating it.
// is there something faster I can do?
func parseEntireFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	p := parser.New()
	doc := p.Parse(content)

	var buffer bytes.Buffer
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			if text, ok := node.(*ast.Text); ok {
				buffer.WriteString(string(text.Leaf.Literal))
			}

			if link, ok := node.(*ast.Link); ok {
				buffer.WriteString(string(link.Destination))
			}

			if code, ok := node.(*ast.CodeBlock); ok {
				buffer.WriteString(string(code.Literal))
			}

			if h, ok := node.(*ast.HTMLBlock); ok {
				buffer.WriteString(string(h.Leaf.Literal))
			}
		}

		return ast.GoToNext
	})

	return buffer.String(), nil
}

// TODO: How can I narrow TokenType to this list of consts?
const (
	EOF     = "EOF"
	ILLEGAL = "ILLEGAL"
	WORD    = "WORD"
)

type TokenType string

type Token struct {
	tokenType TokenType
	literal   string
}

func newToken(tokenType TokenType, ch byte) Token {
	return Token{tokenType: tokenType, literal: string(ch)}
}

type Lexer struct {
	content       *string
	cursor        int // current cursor pos
	nextCursorPos int
	ch            byte // current char being read
}

func (l *Lexer) skipWhiteSpace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isNumber(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

type TermFreqLog struct {
	term string
	freq int
}

func getStats(collection TermFreq) []TermFreqLog {
	out := make([]TermFreqLog, 0, len(collection))
	for t, f := range collection {
		out = append(out, TermFreqLog{t, f})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].freq > out[j].freq
	})

	return out
}

func logTopTerms(filePath string, stats []TermFreqLog, threshold int) {
	fmt.Println(filePath)
	for i, tf := range stats {
		if i < threshold {
			fmt.Println("  ", tf.term, "=>", tf.freq)
		} else {
			break
		}
	}
}

func (l *Lexer) nextToken() Token {
	l.skipWhiteSpace()

	var tok Token
	switch l.ch {
	case 0:
		tok.literal = ""
		tok.tokenType = EOF
	default:
		if isLetter(l.ch) || isNumber(l.ch) {
			tok.literal = strings.ToUpper(l.readWord())
			tok.tokenType = WORD
			return tok
		} else {
			tok = newToken("ILLEGAL", l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readChar() {
	if l.nextCursorPos >= len(*l.content) {
		l.ch = 0
	} else {
		l.ch = (*l.content)[l.nextCursorPos]
	}

	l.cursor = l.nextCursorPos
	l.nextCursorPos += 1
}

func (l *Lexer) readWord() string {
	wordStart := l.cursor

	for isLetter(l.ch) || isNumber(l.ch) {
		l.readChar()
	}

	return (*l.content)[wordStart:l.cursor]
}

func New(input string) *Lexer {
	l := &Lexer{content: &input}
	l.readChar()
	return l
}

type TermFreq map[string]int
type TermFreqIndex map[string]TermFreq

type application struct {
	dirPath       string
	indexPath     string
	staticContent string
}

func (app *application) index() TermFreqIndex {
	dirList, err := os.ReadDir(app.dirPath)

	if err != nil {
		println("ERROR: could not read directory", app.dirPath, err.Error())
		os.Exit(1)
	}

	termFreqIndex := make(TermFreqIndex)

	for _, dir := range dirList {
		if !dir.IsDir() {
			fullPath := app.dirPath + dir.Name()

			fmt.Println("Indexing", fullPath+"...")

			content, err := parseEntireFile(fullPath)

			if err != nil {
				// Log the error and continue
				println("ERROR: could not read file", dir.Name(), ":", err.Error())
				continue
			}

			lexer := New(content)
			tf := make(TermFreq)

			for tok := lexer.nextToken(); tok.tokenType != EOF; tok = lexer.nextToken() {
				if tok.tokenType == WORD {
					if _, ok := tf[tok.literal]; ok {
						tf[tok.literal] += 1
					} else {
						tf[tok.literal] = 1
					}
				}
			}

			termFreqIndex[fullPath] = tf
		}
	}

	fmt.Println("Saving", app.indexPath+"...")
	file, _ := json.MarshalIndent(termFreqIndex, "", "")

	err = os.WriteFile(app.indexPath, file, 0666)
	if err != nil {
		println("ERROR: could not write file", app.indexPath, ":", err.Error())
	}

	return termFreqIndex
}

func (app *application) search(query string) {
	if _, err := os.Stat(app.indexPath); err == nil {
		indexFile, err := os.ReadFile(app.indexPath)

		if err != nil {
			fmt.Println("ERROR: could not open saved index", app.indexPath, err.Error())
		}

		var termFreqIndex TermFreqIndex
		json.Unmarshal(indexFile, &termFreqIndex)
		corpusNumber := len(termFreqIndex)
		queryLexer := New(query)

		for tok := queryLexer.nextToken(); tok.tokenType != EOF; tok = queryLexer.nextToken() {
			if tok.tokenType == WORD {
				termDocCount := 0

				// Count occurances of the term in the entire document corpus
				for _, tf := range termFreqIndex {
					for t := range tf {
						if t == tok.literal {
							termDocCount += 1
						}
					}
				}

				for doc, tf := range termFreqIndex {
					termTotal := 0
					termFreq := 0

					for t, f := range tf {
						termTotal += f

						if t == tok.literal {
							termFreq = f
						}
					}

					idf := math.Log10(float64(corpusNumber) / (float64(termDocCount) + 1))
					tf := float64(termFreq)
					tfidf := float64(tf) * idf

					fmt.Println(doc)
					fmt.Println(" ", tok.literal, "=>", "tf:", tf)
					fmt.Println(" ", tok.literal, "=>", "idf:", idf)
					fmt.Println(" ", tok.literal, "=>", "tfidf:", tfidf)
				}
			}
		}
	}

}

func (app *application) serve() {
  http.Handle("/", http.FileServer(http.Dir(app.staticContent)))
  fmt.Println("serving on port :3000")
  err := http.ListenAndServe(":3000", nil)
  if err != nil {
    fmt.Println("Error running serve subCommand", err.Error())
  }
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter a subCommand [index, search, serve]")
		os.Exit(1)
	}

	app := &application{dirPath: "content/craftinginterpreters/book/", indexPath: "index.json", staticContent: "./static"}
	subCommand := os.Args[1]

	switch subCommand {
	case "index":
		app.index()
	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Please enter a search term")
			os.Exit(1)
		}

		query := os.Args[2]
		app.search(query)
	case "serve":
    app.serve()
	default:
		fmt.Println("Sub-Command not supported")
		os.Exit(1)
	}
}
