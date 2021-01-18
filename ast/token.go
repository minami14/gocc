package ast

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"unicode"
)

const (
	ReservedToken = iota
	NumberToken
	EOFToken

	UnexpectToken = -1
)

type TKind int

var Kinds = map[string]TKind{
	"+": ReservedToken,
	"-": ReservedToken,
	"*": ReservedToken,
	"/": ReservedToken,
	"(": ReservedToken,
	")": ReservedToken,
}

func IsNum(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func Kind(s string) (TKind, error) {
	if IsNum(s) {
		return NumberToken, nil
	}

	k, ok := Kinds[s]
	if !ok {
		return UnexpectToken, fmt.Errorf("unexpected token: %v", s)
	}
	return k, nil
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

	for i := 0; i < len(data); i++ {
		if data[i] == ' ' {
			continue
		}

		k, err := Kind(string(data[i]))
		if err != nil {
			return nil, err
		}
		switch k {
		case ReservedToken:
			tok := &Token{
				Kind:   ReservedToken,
				String: string(data[i]),
			}
			current.Next = tok
			current = tok
			continue
		case NumberToken:
			h := i
			for ; i < len(data)-1; i++ {
				if !unicode.IsNumber(rune(data[i+1])) {
					break
				}
			}
			str := string(data[h : i+1])
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
			continue
		}
	}
	current.Next = &Token{Kind: EOFToken}
	return head.Next, nil
}
