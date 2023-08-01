package lexer

import (
  "strings"

  "gosearch/token"
)


func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{TokenType: tokenType, Literal: string(ch)}
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


func (l *Lexer) NextToken() token.Token {
	l.skipWhiteSpace()

	var tok token.Token
	switch l.ch {
	case 0:
		tok.Literal = ""
		tok.TokenType = token.EOF
	default:
		if isLetter(l.ch) || isNumber(l.ch) {
			tok.Literal = strings.ToUpper(l.readWord())
			tok.TokenType = token.WORD
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
