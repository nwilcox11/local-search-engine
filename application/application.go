package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"

	"gosearch/lexer"
	"gosearch/token"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

type TermFreqIndex map[string]*TermFreqStruct

type TermFreq map[string]int

type TermFreqStruct struct {
	Meta string
	Idx  TermFreq
}

type Application struct {
	DirPath       string
	IndexPath     string
	StaticContent string
}

const smallestDocLength = 34
const nextSmallestDocLength = 75
const domain = "craftinginterpreters.com/"

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

func toHtmlType(path string) string {
	splitPath := strings.Split(path, ".")
	return splitPath[0] + ".html"
}

// lol
func contentPreview(content string) string {
	l := len(content)
  var out string
	if l <= smallestDocLength+1 {
		out = content[:smallestDocLength]
	} else if l <= nextSmallestDocLength {
		out = content[:nextSmallestDocLength-5]
	} else {
    out = content[:200]
  }

	return strings.ReplaceAll(out, "\n", " ")
}

func (app *Application) Index() TermFreqIndex {
	dirList, err := os.ReadDir(app.DirPath)

	if err != nil {
		println("ERROR: could not read directory", app.DirPath, err.Error())
		os.Exit(1)
	}

	termFreqIndex := make(TermFreqIndex)

	for _, dir := range dirList {
		if !dir.IsDir() {
			fullPath := app.DirPath + dir.Name()
			indexKey := domain + toHtmlType(dir.Name())

			fmt.Println("Indexing", fullPath+"...")

			content, err := parseEntireFile(fullPath)

			if err != nil {
				// Log the error and continue
				println("ERROR: could not read file", dir.Name(), ":", err.Error())
				continue
			}

			lexer := lexer.New(content)
			tf := make(TermFreq)

			for tok := lexer.NextToken(); tok.TokenType != token.EOF; tok = lexer.NextToken() {
				if tok.TokenType == token.WORD {
					if _, ok := tf[tok.Literal]; ok {
						tf[tok.Literal] += 1
					} else {
						tf[tok.Literal] = 1
					}
				}
			}

			termFreqContainer := &TermFreqStruct{Meta: contentPreview(content), Idx: tf}
			termFreqIndex[indexKey] = termFreqContainer
		}
	}

	fmt.Println("Saving", app.IndexPath+"...")
	file, _ := json.MarshalIndent(termFreqIndex, "", "")

	err = os.WriteFile(app.IndexPath, file, 0666)
	if err != nil {
		println("ERROR: could not write file", app.IndexPath, ":", err.Error())
	}

	return termFreqIndex
}

type tfidfTermDoc struct {
	Term  string `json:"term"`
	Doc   string `json:"doc"`
	tf    float64
	Idf   float64 `json:"idf"`
	Tfidf float64 `json:"tfidf"`
	Meta  string
}
type tfidfIndexResult = map[string][]tfidfTermDoc

func (app *Application) Search(query string) (tfidfIndexResult, error) {
	var out tfidfIndexResult
	if _, err := os.Stat(app.IndexPath); err == nil {
		indexFile, err := os.ReadFile(app.IndexPath)

		if err != nil {
			fmt.Println("ERROR: could not open saved index", app.IndexPath, err.Error())
			os.Exit(1)
			return nil, err
		}

		var termFreqIndex TermFreqIndex
		json.Unmarshal(indexFile, &termFreqIndex)
		corpusNumber := len(termFreqIndex)
		queryLexer := lexer.New(query)
		out = make(tfidfIndexResult)

		for tok := queryLexer.NextToken(); tok.TokenType != token.EOF; tok = queryLexer.NextToken() {
			if tok.TokenType == token.WORD {
				termDocCount := 0

				// Count occurances of the term in the entire document corpus
				for _, tf := range termFreqIndex {
					for t := range tf.Idx {
						if t == tok.Literal {
							termDocCount += 1
						}
					}
				}

				// Skip WORD if it is not found in the document.
				if termDocCount == 0 {
					continue
				}

				out[tok.Literal] = make([]tfidfTermDoc, 0, corpusNumber)

				for doc, tfs := range termFreqIndex {
					termTotal := 0
					termFreq := 0

					for t, f := range tfs.Idx {
						termTotal += f

						if t == tok.Literal {
							termFreq = f
						}
					}

					idf := math.Log10(float64(corpusNumber) / (float64(termDocCount) + 1))
					tf := float64(termFreq)
					tfidf := float64(tf) * idf

					if _, ok := out[tok.Literal]; ok && tfidf > 0 {
						out[tok.Literal] = append(out[tok.Literal], tfidfTermDoc{Doc: doc, Meta: tfs.Meta, Idf: idf, Tfidf: tfidf, Term: tok.Literal})
					}
				}
			}

			sort.Slice(out[tok.Literal], func(i, j int) bool {
				return out[tok.Literal][i].Tfidf > out[tok.Literal][j].Tfidf
			})
		}
	}
	return out, nil
}

func (app *Application) Serve() {
	http.Handle("/", http.FileServer(http.Dir(app.StaticContent)))

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")

		result, _ := app.Search(q)
		body, _ := json.Marshal(result)

		w.Header().Add("Content-Type", "application/json")

		w.Write(body)
	})

	fmt.Println("serving on port :3000")
	err := http.ListenAndServe(":3000", nil)

	if err != nil {
		fmt.Println("Error running serve subCommand", err.Error())
	}
}
