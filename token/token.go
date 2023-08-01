package token

// TODO: How can I narrow TokenType to this list of consts?
const (
	EOF     = "EOF"
	ILLEGAL = "ILLEGAL"
	WORD    = "WORD"
)

type TokenType string

type Token struct {
	TokenType TokenType
	Literal   string
}
