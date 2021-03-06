package po

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
)

// scanner scans the message fields of a po file.
// it is a mirror of the writer.
type scanner struct {
	*bufio.Scanner
	hasNext bool
	err     error
}

func newScanner(r io.Reader) *scanner {
	return &scanner{bufio.NewScanner(r), true, nil}
}

// nextmsg goes to the next message, skipping blank lines in between.
func (s *scanner) nextmsg() bool {
	for {
		if s.err != nil {
			return false
		}
		if !s.Scan() {
			return false
		}
		// skip newlines and lines that are precisely "#"
		b := s.Bytes()
		if len(bytes.TrimSpace(b)) > 1 {
			return true
		}
	}
}

func (s *scanner) mul(prefix string) []string {
	var r []string
	for s.prefix(prefix) {
		r = append(r, s.txt(prefix))
		if !s.Scan() {
			break
		}
	}
	return r
}

func (s *scanner) spc(prefix string) []string {
	var r []string
	if s.prefix(prefix) {
		r = append(r, strings.Fields(s.txt(prefix))...)
		s.Scan()
	}
	return r
}

func (s *scanner) one(prefix string) string {
	var r string
	if s.prefix(prefix) {
		r = s.txt(prefix)
		s.Scan()
	}
	return r
}

// quo reads a quoted string after the given prefix.
// multiline strings are handled.
func (s *scanner) quo(prefix string) string {
	var r string
	if s.prefix(prefix) {
		r = s.unquote(s.txt(prefix))
		for {
			if !s.Scan() {
				return r
			}
			if len(s.Bytes()) > 0 && s.Bytes()[0] == '"' {
				r += s.unquote(s.Text())
				continue
			}
			break
		}
	}
	return r
}

// msgstr parses the msgstr section of a message record.
// it handles multiline messages as well as indexed plural forms.
func (s *scanner) msgstr() []string {
	if s.prefix("msgstr ") {
		return []string{s.quo("msgstr ")}
	}

	var r []string
	for {
		var prefix = "msgstr[" + strconv.Itoa(len(r)) + "] "
		if !s.prefix(prefix) {
			return r
		}
		r = append(r, s.quo(prefix))
	}
}

func (s *scanner) unquote(str string) string {
	var r, err = strconv.Unquote(str)
	if err != nil {
		s.err = err
	}
	return r
}

// Err returns the last error encountered, if any.
func (s *scanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.Scanner.Err()
}

// txt returns the text on the current line after the given prefix, trimming space.
func (s *scanner) txt(prefix string) string {
	return strings.TrimSpace(s.Text()[len(prefix):])
}

// prefix returns true if the current line begins with the given prefix.
func (s *scanner) prefix(prefix string) bool {
	return bytes.HasPrefix(s.Bytes(), []byte(prefix))
}
