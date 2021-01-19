package ast

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"
)

const (
	ReservedToken = iota
	IdentToken
	NumberToken
	ReturnToken
	IfToken
	ElseToken
	WhileToken
	ForToken
	EOFToken
)

type TKind int

var TKinds = map[string]TKind{
	"+":  ReservedToken,
	"-":  ReservedToken,
	"*":  ReservedToken,
	"/":  ReservedToken,
	"%":  ReturnToken,
	"(":  ReservedToken,
	")":  ReservedToken,
	"==": ReservedToken,
	"!=": ReservedToken,
	"<=": ReservedToken,
	">=": ReservedToken,
	"<":  ReservedToken,
	">":  ReservedToken,
	"=":  ReservedToken,
	";":  ReservedToken,
	"{":  ReturnToken,
	"}":  ReturnToken,
}

func IsNum(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func tokenIs(p []byte, s string) bool {
	if string(p) == s {
		return true
	}

	if len(p) < len(s) {
		return false
	}

	if string(p[:len(s)]) != s {
		return false
	}

	v := p[len(s)]
	isSmall := v >= 'a' && v <= 'z'
	isLarge := v >= 'A' && v <= 'Z'
	isNum := v >= '0' && v <= '9'
	return !(isSmall || isLarge || isNum || v == '_')
}

func TokenKind(p []byte) (TKind, int) {
	for i := 0; i < len(p); i++ {
		if !IsNum(string(p[i])) {
			if i == 0 {
				break
			}
			return NumberToken, i
		}
		if i == len(p)-1 {
			return NumberToken, i + 1
		}
	}

	if tokenIs(p, "return") {
		return ReturnToken, 6
	}

	if tokenIs(p, "if") {
		return IfToken, 2
	}

	if tokenIs(p, "else") {
		return ElseToken, 4
	}

	if tokenIs(p, "while") {
		return WhileToken, 5
	}

	if tokenIs(p, "for") {
		return ForToken, 3
	}

	if len(p) >= 2 {
		k, ok := TKinds[string(p[:2])]
		if ok {
			return k, 2
		}
	}

	k, ok := TKinds[string(p[:1])]
	if ok {
		return k, 1
	}

	for i, v := range p {
		isSmall := v >= 'a' && v <= 'z'
		isLarge := v >= 'A' && v <= 'Z'
		isNum := v >= '0' && v <= '9'
		if !(isSmall || isLarge || isNum || v == '_') {
			return IdentToken, i
		}
	}
	return IdentToken, len(p)
}

type Token struct {
	Kind   TKind
	Next   *Token
	Value  int
	String string
	Offset int
	Row    int
	Col    int
}

func Tokenize(src io.Reader) (*Token, error) {
	locals := make(map[string]int)
	head := &Token{}
	current := head
	buf, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	d := bytes.Split(buf, []byte{'\n'})

	for row, data := range d {
		for i := 0; i < len(data); {
			if data[i] == ' ' || data[i] == '\n' || data[i] == '\t' {
				i++
				continue
			}

			k, l := TokenKind(data[i:])
			str := string(data[i : i+l])
			tok := &Token{
				Kind:   k,
				String: str,
				Row:    row + 1,
				Col:    i + 1,
			}
			if k == NumberToken {
				tok.Value, err = strconv.Atoi(str)
				if err != nil {
					return nil, err
				}
			}
			if k == IdentToken {
				n, ok := locals[str]
				if !ok {
					n = len(locals) * 8
					locals[str] = n
				}
				tok.Offset = n
			}
			current.Next = tok
			current = tok
			i += l
		}
	}
	current.Next = &Token{Kind: EOFToken}
	return head.Next, nil
}
