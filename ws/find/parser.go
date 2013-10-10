// Copyright 2013 Martin Schnabel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package find

import "fmt"

// The following query spec uses the same notation as the go language spec:
// http://golang.org/doc/spec#Notation
//
//    special_char = "$", "*", "/", `\`
//    char = unicode_char | escaped_char
//    unicode_char = // an arbitrary Unicode code point that is not a special_char
//    escaped_char = `\` special_char
//
//    word = char {char}
//    term = "*" ["*" [words] | words ["*"]] | words ["*" ["*"]]
//    words = word {"*" word}
//    query = ["/"] term {"/" term} ["/"] ["$"]
//
type tok int

const (
	illegal tok = iota
	slash
	star
	dolar
	word
	eof
)

type parser struct {
	raw []rune
	buf []rune
	i   int
	q   Query
}

func (p *parser) parse(input string) (Query, error) {
	p.raw, p.i = []rune(input), 0
	q := Query{Raw: input}
	tok := p.next()
	// leading slash signifies an absolute path
	if tok == slash {
		q.Absolute = true
		tok = p.next()
	}
	var cur Term
Loop:
	for tok != eof {
		switch tok {
		case star:
			cur.Wildcard = Start
			switch tok = p.next(); tok {
			case star:
				cur.Wildcard = DblStart
				if tok = p.next(); tok != word {
					break
				}
				tok = p.parseWords(&cur)
			case word:
				tok = p.parseWords(&cur)
			}
		case word:
			tok = p.parseWords(&cur)
		case dolar:
			break Loop
		}
		if tok == dolar {
			cur.Type = File
			tok = p.next()
		}
		if tok == slash {
			tok = p.next()
			cur.Type = Dir
			if tok == eof {
				break Loop
			}
			q.Terms = append(q.Terms, cur)
			cur = Term{}
			continue
		}
		break
	}
	q.Terms = append(q.Terms, cur)
	if tok != eof {
		return q, fmt.Errorf("unexpected token %d", tok)
	}
	return q, nil
}

func (p *parser) parseWords(cur *Term) tok {
	cur.Words = append(cur.Words, string(p.buf))
	tok := p.next()
	for tok == star {
		if tok = p.next(); tok == word {
			cur.Words = append(cur.Words, string(p.buf))
			tok = p.next()
		} else if cur.Wildcard&DblStart == 0 && tok == star {
			cur.Wildcard |= DblEnd
			tok = p.next()
			break
		} else {
			cur.Wildcard |= End
			break
		}
	}
	return tok
}

func (p *parser) next() tok {
	if p.i >= len(p.raw) {
		return eof
	}
	var n rune
	n, p.i = p.raw[p.i], p.i+1
	switch n {
	case '/':
		return slash
	case '*':
		return star
	case '$':
		return dolar
	}
	p.buf = p.buf[:0]
	for {
		switch n {
		case '/', '*', '$':
			p.i--
			return word
		case '\\':
			if p.i < len(p.raw) {
				n, p.i = p.raw[p.i], p.i+1
			} else {
				n, p.i = rune(0), p.i+1
			}
			switch n {
			case '$', '*', '/', '\\':
				p.buf = append(p.buf, n)
			default:
				return illegal
			}
		default:
			p.buf = append(p.buf, n)
		}
		if p.i >= len(p.raw) {
			break
		}
		n, p.i = p.raw[p.i], p.i+1
	}
	return word
}
