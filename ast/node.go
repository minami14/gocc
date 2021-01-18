package ast

import (
	"fmt"
)

const (
	AddNode = iota
	SubNode
	MulNode
	DivNode
	NumNode
)

type NodeKind int

type Node struct {
	Kind   NodeKind
	Left   *Node
	Right  *Node
	Value  int
	String string
}

func Expr(tok *Token) (*Node, error) {
	node, _, err := expr(tok)
	return node, err
}

func expr(tok *Token) (*Node, *Token, error) {
	var (
		node, right *Node
		err         error
	)
	node, tok, err = mul(tok)
	if err != nil {
		return node, tok, err
	}

	for {
		if tok.Kind != ReservedToken {
			return node, tok, nil
		}

		switch tok.String {
		case "+":
			tok = tok.Next
			right, tok, err = mul(tok)
			if err != nil {
				return node, tok, err
			}
			node = &Node{
				Kind:   AddNode,
				Left:   node,
				Right:  right,
				String: "+",
			}
		case "-":
			tok = tok.Next
			right, tok, err = mul(tok)
			if err != nil {
				return node, tok, err
			}
			node = &Node{
				Kind:   SubNode,
				Left:   node,
				Right:  right,
				String: "-",
			}
		default:
			return node, tok, nil
		}
	}
}

func mul(tok *Token) (*Node, *Token, error) {
	var (
		node, right *Node
		err         error
	)
	node, tok, err = unary(tok)
	if err != nil {
		return node, tok, err
	}

	for {
		switch tok.String {
		case "*":
			tok = tok.Next
			right, tok, err = unary(tok)
			if err != nil {
				return node, tok, err
			}
			node = &Node{
				Kind:   MulNode,
				Left:   node,
				Right:  right,
				String: "*",
			}
		case "/":
			tok = tok.Next
			right, tok, err = unary(tok)
			if err != nil {
				return node, tok, err
			}
			node = &Node{
				Kind:   DivNode,
				Left:   node,
				Right:  right,
				String: "/",
			}
		default:
			return node, tok, nil
		}
	}
}

func primary(tok *Token) (*Node, *Token, error) {
	var (
		node *Node
		err  error
	)
	if tok.String == "(" {
		tok = tok.Next
		node, tok, err = expr(tok)
		if err != nil {
			return node, tok, err
		}
		if tok.String != ")" {
			return node, tok, fmt.Errorf("unexpected token: %v", tok.String)
		}
		tok = tok.Next
		return node, tok, err
	}

	if tok.Kind != NumberToken {
		return node, tok, fmt.Errorf("unexpected token: %v", tok.String)
	}

	node = &Node{
		Kind:   NumNode,
		Value:  tok.Value,
		String: tok.String,
	}
	tok = tok.Next

	return node, tok, err
}

func unary(tok *Token) (*Node, *Token, error) {
	switch tok.String {
	case "+":
		node, tok, err := primary(tok.Next)
		return node, tok, err
	case "-":
		node, tok, err := primary(tok.Next)
		node = &Node{
			Kind: SubNode,
			Left: &Node{
				Kind:   NumNode,
				Left:   nil,
				Right:  nil,
				Value:  0,
				String: "0",
			},
			Right:  node,
			Value:  0,
			String: "-",
		}
		return node, tok, err
	default:
		node, tok, err := primary(tok)
		return node, tok, err
	}
}

func (n *Node) Gen() ([]byte, error) {
	var buf []byte
	if n.Kind == NumNode {
		buf = []byte("  push ")
		buf = append(buf, n.String...)
		buf = append(buf, '\n')
		return buf, nil
	}

	if n.Left != nil {
		l, err := n.Left.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, l...)
	}

	if n.Right != nil {
		r, err := n.Right.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, r...)
	}

	buf = append(buf, "  pop rdi\n  pop rax\n"...)

	switch n.Kind {
	case AddNode:
		buf = append(buf, "  add rax, rdi\n"...)
	case SubNode:
		buf = append(buf, "  sub rax, rdi\n"...)
	case MulNode:
		buf = append(buf, "  imul rax, rdi\n"...)
	case DivNode:
		buf = append(buf, "  cqo\n  idiv rax, rdi\n"...)
	}

	buf = append(buf, "  push rax\n"...)
	return buf, nil
}
