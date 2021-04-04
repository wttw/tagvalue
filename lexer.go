package tagvalue

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"html/template"
	"strings"
	"unicode/utf8"
)

/*
From RFC 6376:

   tag-list  =  tag-spec *( ";" tag-spec ) [ ";" ]
   tag-spec  =  [FWS] tag-name [FWS] "=" [FWS] tag-value [FWS]
   tag-name  =  ALPHA *ALNUMPUNC
   tag-value =  [ tval *( 1*(WSP / FWS) tval ) ]
                     ; Prohibits WSP and FWS at beginning and end
   tval      =  1*VALCHAR
   VALCHAR   =  %x21-3A / %x3C-7E
                     ; EXCLAMATION to TILDE except SEMICOLON
   ALNUMPUNC =  ALPHA / DIGIT / "_"

    WSP =   SP / HTAB
    FWS =   [*WSP CRLF] 1*WSP
*/

type itemType int

const (
	itemError itemType = iota // error occurred;
	itemEOF
	itemTag
	itemValue
)

const eof = -1

// item represents a token returned from the scanner.
type item struct {
	typ itemType  // Type, such as itemNumber.
	val string    // Value, such as "23.2".
	start int     // Offset of the beginning of this token
}

type lexer struct {
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
}

type Item struct {
	Tag string
	Value string
	TagPos int
	ValuePos int
}

type ParseError struct {
	Message string
	Pos int
}

func (e ParseError) Error() string {
	return e.Message
}

type FieldError struct {
	Message template.HTML
	Severity string
}

type Field struct {
	Item
	Index     int
	Duplicate bool
	Defined bool
	Errors    []FieldError
}

func (f *Field) annotate(severity string, message interface{}, data ...interface{}) {
	var anno template.HTML
	switch v := message.(type) {
	case string:
		anno = template.HTML(template.HTMLEscapeString(v))
	case template.HTML:
		anno = v
	case template.Template:
		var buff bytes.Buffer
		err := v.Execute(&buff, data)
		if err != nil {
			anno = template.HTML(template.HTMLEscapeString(fmt.Sprintf("Internal error in %s: %v", v.Name(), err)))
		} else {
			anno = template.HTML(buff.String())
		}
	default:
		anno = template.HTML(fmt.Sprintf("Internal error, annotation type %T: %s", message, template.HTMLEscaper(data)))
	}
	f.Errors = append(f.Errors, FieldError{Message: anno, Severity: severity})
}

func (f *Field) addError(message interface{}, data ...interface{}) {
	f.annotate("danger", message, data)
}

func (f *Field) addWarning(message interface{}, data ...interface{}) {
	f.annotate("warning", message, data)
}

func (f *Field) addInfo(message interface{}, data ...interface{}) {
	f.annotate("info", message, data)
}


// NewMap parses a rfc 6376 tag=value input and returns
// a map of tag to item
func NewMap(input string) (map[string]Field, error) {
	fields := map[string]Field{}
	items, err := NewTagValue(input)
	if err != nil {
		return nil, err
	}
	for i, item := range items {
		_, seen := fields[item.Tag]
		v := Field{
			Item: item,
			Index: i,
			Duplicate: seen,
			Defined: true,
			Errors: []FieldError{},
		}
		fields[item.Tag] = v
	}
	return fields, nil
}

// NewTagValue parses a rfc 6376 tag=value input and returns
// a slice containing all the parsed tag-value pairs
func NewTagValue(input string) ([]Item, error) {
	l := &lexer{
		input:input,
	}

	items := []Item{}

	for {
		var item Item
		err := l.skipOptionalFws()
		if err != nil {
			return nil, err
		}

		// tag
		r := l.next()
		if r == eof {
			return items, nil
		}
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return nil, ParseError{"expecting alpha character in tag", l.pos}
		}
		l.acceptTagRun()
		item.Tag = l.input[l.start:l.pos]
		item.TagPos = l.start

		// =
		err = l.skipOptionalFws()
		if err != nil {
			return nil, err
		}

		r = l.next()
		if r != '=' {
			return nil, ParseError{"expecting '='", l.pos}
		}
		err = l.skipOptionalFws()
		if err != nil {
			return nil, err
		}

		// value
		value:
		for {
			for {
				r = l.next()
				if r != ';' && r >= '!' && r <= '~' {
					continue
				}
				l.backup()
				whitespacePos := l.pos
				err = l.acceptOptionalFws()
				if err != nil {
					return nil, err
				}
				r = l.next()
				if r == eof || r == ';' {
					item.Value = l.input[l.start:whitespacePos]
					item.ValuePos = l.start
					items = append(items, item)
					break value
				}
			}
		}
	}
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width =
		utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}
// backup steps back one rune.
// Can be called only once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// accept consumes the next rune
// if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}
// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// acceptTagRun consumes alphamerics plus underscore
func (l *lexer) acceptTagRun() {
	l.acceptRun("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_")
}

// acceptOptionalFws consumes an optional folding whitespace
func (l *lexer) acceptOptionalFws() error {
	l.acceptRun(" \t")
	if strings.HasPrefix(l.input[l.pos:], "\r\n") {
		l.pos += 2
		if l.accept(" \t") {
			l.acceptRun(" \t")
			return nil
		}
		l.backup()
		return errors.New("malformed folding whitespace")
	}
	return nil
}

func (l *lexer) skipOptionalFws() error {
	err := l.acceptOptionalFws()
	if err != nil {
		return err
	}
	l.ignore()
	return nil
}
