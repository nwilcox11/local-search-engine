package application

import (
  "bytes"
  "fmt"
  "net/http"
  "os"
  "sort"
  "math"
	"encoding/json"

  "gosearch/lexer"
  "gosearch/token"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

type TermFreq map[string]int
type TermFreqIndex map[string]TermFreq

type Application struct {
	DirPath       string
	IndexPath     string
	StaticContent string
}

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

type TermFreqLog struct {
	term string
	freq int
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

			termFreqIndex[fullPath] = tf
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

func (app *Application) Search(query string) {
	if _, err := os.Stat(app.IndexPath); err == nil {
		indexFile, err := os.ReadFile(app.IndexPath)

		if err != nil {
			fmt.Println("ERROR: could not open saved index", app.IndexPath, err.Error())
		}

		var termFreqIndex TermFreqIndex
		json.Unmarshal(indexFile, &termFreqIndex)
		corpusNumber := len(termFreqIndex)
		queryLexer := lexer.New(query)

		for tok := queryLexer.NextToken(); tok.TokenType != token.EOF; tok = queryLexer.NextToken() {
			if tok.TokenType == token.WORD {
				termDocCount := 0

				// Count occurances of the term in the entire document corpus
				for _, tf := range termFreqIndex {
					for t := range tf {
						if t == tok.Literal {
							termDocCount += 1
						}
					}
				}

				for doc, tf := range termFreqIndex {
					termTotal := 0
					termFreq := 0

					for t, f := range tf {
						termTotal += f

						if t == tok.Literal {
							termFreq = f
						}
					}

					idf := math.Log10(float64(corpusNumber) / (float64(termDocCount) + 1))
					tf := float64(termFreq)
					tfidf := float64(tf) * idf

					fmt.Println(doc)
					fmt.Println(" ", tok.Literal, "=>", "tf:", tf)
					fmt.Println(" ", tok.Literal, "=>", "idf:", idf)
					fmt.Println(" ", tok.Literal, "=>", "tfidf:", tfidf)
				}
			}
		}
	}

}

func (app *Application) Serve() {
  http.Handle("/", http.FileServer(http.Dir(app.StaticContent)))
  fmt.Println("serving on port :3000")
  err := http.ListenAndServe(":3000", nil)
  if err != nil {
    fmt.Println("Error running serve subCommand", err.Error())
  }
}
