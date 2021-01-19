package ast

import (
	"io"
	"io/ioutil"
	"strconv"
)

const (
	ReservedToken = iota
	IdentToken
	NumberToken
	ReturnToken
	EOFToken

	UnexpectToken = -1
)

type TKind int

var TKinds = map[string]TKind{
	"+":  ReservedToken,
	"-":  ReservedToken,
	"*":  ReservedToken,
	"/":  ReservedToken,
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
}

func IsNum(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func TokenKind(p []byte) (TKind, int, error) {
	for i := 0; i < len(p); i++ {
		if !IsNum(string(p[i])) {
			if i == 0 {
				break
			}
			return NumberToken, i, nil
		}
		if i == len(p)-1 {
			return NumberToken, i + 1, nil
		}
	}

	returnLen := len("return")
	if len(p) >= returnLen {
		if string(p[:returnLen]) == "return" {
			if len(p) > returnLen {
				if v := p[returnLen]; (v >= 'a' && v <= 'z') || (v >= 'A' && v <= 'Z') || (v >= '0' && v <= '9') || v == '_' {
					for i := returnLen; i < len(p); i++ {
						if p[i] == ' ' || p[i] == '\n' {
							return LVarNode, i, nil
						}
						if i == len(p)-1 {
							return LVarNode, i + 1, nil
						}
					}
				}
			}
			return ReturnToken, returnLen, nil
		}
	}

	if len(p) >= 2 {
		k, ok := TKinds[string(p[:2])]
		if ok {
			return k, 2, nil
		}
	}

	k, ok := TKinds[string(p[:1])]
	if ok {
		return k, 1, nil
	}

	return IdentToken, 1, nil
}

type Token struct {
	Kind   TKind
	Next   *Token
	Value  int
	String string
}

func Tokenize(src io.Reader) (*Token, error) {
	head := &Token{}
	current := head
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(data); {
		if data[i] == ' ' || data[i] == '\n' {
			i++
			continue
		}

		k, l, err := TokenKind(data[i:])
		if err != nil {
			return nil, err
		}
		str := string(data[i : i+l])
		switch k {
		case ReservedToken:
			tok := &Token{
				Kind:   ReservedToken,
				String: str,
			}
			current.Next = tok
			current = tok
			i += l
			continue
		case IdentToken:
			tok := &Token{
				Kind:   IdentToken,
				String: str,
			}
			current.Next = tok
			current = tok
			i += l
			continue
		case NumberToken:
			val, err := strconv.Atoi(str)
			if err != nil {
				return nil, err
			}
			tok := &Token{
				Kind:   NumberToken,
				Value:  val,
				String: str,
			}
			current.Next = tok
			current = tok
			i += l
			continue
		case ReturnToken:
			tok := &Token{
				Kind:   ReturnToken,
				String: str,
			}
			current.Next = tok
			current = tok
			i += l
			continue
		}
	}
	current.Next = &Token{Kind: EOFToken}
	return head.Next, nil
}
