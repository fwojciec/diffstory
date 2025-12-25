// Package chroma provides syntax highlighting using the chroma library.
package chroma

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/fwojciec/diffview"
)

// Compile-time interface verification.
var _ diffview.Tokenizer = (*Tokenizer)(nil)

// Tokenizer extracts syntax tokens using chroma.
type Tokenizer struct{}

// NewTokenizer creates a new chroma-based tokenizer.
func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

// Tokenize splits source code into syntax-highlighted tokens for the given language.
// Returns nil if the language is not supported or an error occurs.
// Returns an empty slice for empty source (valid input, no tokens).
func (t *Tokenizer) Tokenize(language, source string) []diffview.Token {
	if source == "" {
		return []diffview.Token{}
	}

	lexer := lexers.Get(language)
	if lexer == nil {
		return nil
	}

	// Coalesce for better performance with consecutive tokens of the same type
	lexer = chroma.Coalesce(lexer)

	iterator, err := lexer.Tokenise(nil, source)
	if err != nil {
		return nil
	}

	var tokens []diffview.Token
	for token := iterator(); token != chroma.EOF; token = iterator() {
		style := tokenStyle(token.Type)
		tokens = append(tokens, diffview.Token{
			Text:  token.Value,
			Style: style,
		})
	}

	return tokens
}

// tokenStyle returns the visual style for a chroma token type.
// Colors are loosely based on the One Dark theme.
func tokenStyle(tt chroma.TokenType) diffview.Style {
	// Use direct type comparison for specific types,
	// then fall through to category checks for broader matches.
	switch tt {
	// Keywords
	case chroma.Keyword, chroma.KeywordConstant, chroma.KeywordDeclaration,
		chroma.KeywordNamespace, chroma.KeywordPseudo, chroma.KeywordReserved,
		chroma.KeywordType:
		return diffview.Style{Foreground: "#c678dd", Bold: true}

	// Comments
	case chroma.Comment, chroma.CommentHashbang, chroma.CommentMultiline,
		chroma.CommentPreproc, chroma.CommentPreprocFile, chroma.CommentSingle,
		chroma.CommentSpecial:
		return diffview.Style{Foreground: "#5c6370"}

	// Strings (String* and LiteralString* are aliases, so only use one set)
	case chroma.String, chroma.StringAffix, chroma.StringBacktick, chroma.StringChar,
		chroma.StringDelimiter, chroma.StringDoc, chroma.StringDouble,
		chroma.StringEscape, chroma.StringHeredoc, chroma.StringInterpol,
		chroma.StringOther, chroma.StringRegex, chroma.StringSingle,
		chroma.StringSymbol:
		return diffview.Style{Foreground: "#98c379"}

	// Numbers (Number* and LiteralNumber* are aliases, so only use one set)
	case chroma.Number, chroma.NumberBin, chroma.NumberFloat, chroma.NumberHex,
		chroma.NumberInteger, chroma.NumberIntegerLong, chroma.NumberOct:
		return diffview.Style{Foreground: "#d19a66"}

	// Operators
	case chroma.Operator, chroma.OperatorWord:
		return diffview.Style{Foreground: "#56b6c2"}

	// Builtin names (e.g., println, len, make)
	case chroma.NameBuiltin, chroma.NameBuiltinPseudo:
		return diffview.Style{Foreground: "#e5c07b"}

	// Function names
	case chroma.NameFunction, chroma.NameFunctionMagic:
		return diffview.Style{Foreground: "#61afef"}

	// Other names (general identifiers)
	case chroma.Name, chroma.NameAttribute, chroma.NameClass, chroma.NameConstant,
		chroma.NameDecorator, chroma.NameEntity, chroma.NameException,
		chroma.NameLabel, chroma.NameNamespace, chroma.NameOther,
		chroma.NameProperty, chroma.NameTag, chroma.NameVariable,
		chroma.NameVariableAnonymous, chroma.NameVariableClass,
		chroma.NameVariableGlobal, chroma.NameVariableInstance,
		chroma.NameVariableMagic:
		return diffview.Style{Foreground: "#e06c75"}

	default:
		return diffview.Style{}
	}
}
