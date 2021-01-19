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
	EqNode
	NeNode
	LtNode
	LeNode
	AssignNode
	LVarNode
)

type NKind int

type Node struct {
	Kind   NKind
	Left   *Node
	Right  *Node
	Offset int
	Value  int
	String string
}

func Parse(tok *Token) ([]*Node, error) {
	node, _, err := program(tok)
	return node, err
}

func program(tok *Token) ([]*Node, *Token, error) {
	var (
		nodes []*Node
		node  *Node
		err   error
	)
	for tok != nil && tok.Kind != EOFToken {
		node, tok, err = stmt(tok)
		if err != nil {
			return nodes, tok, err
		}
		nodes = append(nodes, node)
	}
	return nodes, tok, nil
}

func stmt(tok *Token) (*Node, *Token, error) {
	var (
		node *Node
		err  error
	)
	node, tok, err = expr(tok)
	if err != nil {
		return node, tok, err
	}
	if tok.String != ";" && tok.String != "\n" && tok.Kind != EOFToken {
		return node, tok, fmt.Errorf("unexpected token: %v", tok.String)
	}
	tok = tok.Next
	return node, tok, nil
}

func assign(tok *Token) (*Node, *Token, error) {
	var (
		node, right *Node
		err         error
	)
	node, tok, err = equality(tok)
	if err != nil {
		return node, tok, err
	}
	if tok.String == "=" {
		tok = tok.Next
		right, tok, err = assign(tok)
		if err != nil {
			return node, tok, err
		}
		node = &Node{
			Kind:   AssignNode,
			Left:   node,
			Right:  right,
			String: "=",
		}
	}
	return node, tok, err
}

func equality(tok *Token) (*Node, *Token, error) {
	var (
		node, right *Node
		err         error
	)
	node, tok, err = relational(tok)
	for {
		str := tok.String
		if str == "==" || str == "!=" {
			tok = tok.Next
			right, tok, err = relational(tok)
			if err != nil {
				return node, tok, err
			}
			var k NKind
			switch str {
			case "==":
				k = EqNode
			case "!=":
				k = NeNode
			}
			node = &Node{
				Kind:   k,
				Left:   node,
				Right:  right,
				String: str,
			}
		} else {
			return node, tok, nil
		}
	}
}

func relational(tok *Token) (*Node, *Token, error) {
	var (
		node, right *Node
		err         error
	)
	node, tok, err = add(tok)
	if err != nil {
		return node, tok, err
	}
	for {
		str := tok.String
		if str == "<" || str == "<=" || str == ">" || str == ">=" {
			tok = tok.Next
			right, tok, err = add(tok)
			if err != nil {
				return node, tok, err
			}
			var k NKind
			switch str {
			case "<":
				k = LtNode
			case "<=":
				k = LeNode
			case ">":
				k = LtNode
				node, right = right, node
			case ">=":
				k = LeNode
				node, right = right, node
			}
			node = &Node{
				Kind:   k,
				Left:   node,
				Right:  right,
				String: str,
			}
		} else {
			return node, tok, nil
		}
	}
}

func add(tok *Token) (*Node, *Token, error) {
	var (
		node, right *Node
		err         error
	)
	node, tok, err = mul(tok)
	if err != nil {
		return node, tok, err
	}
	for {
		str := tok.String
		if str == "+" || str == "-" {
			tok = tok.Next
			right, tok, err = mul(tok)
			if err != nil {
				return node, tok, err
			}
			var k NKind
			switch str {
			case "+":
				k = AddNode
			case "-":
				k = SubNode
			}
			node = &Node{
				Kind:   k,
				Left:   node,
				Right:  right,
				String: str,
			}
		} else {
			return node, tok, nil
		}
	}
}

func expr(tok *Token) (*Node, *Token, error) {
	return assign(tok)
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
		str := tok.String
		if str == "*" || str == "/" {
			tok = tok.Next
			right, tok, err = unary(tok)
			if err != nil {
				return node, tok, err
			}
			var k NKind
			switch str {
			case "*":
				k = MulNode
			case "/":
				k = DivNode
			}
			node = &Node{
				Kind:   k,
				Left:   node,
				Right:  right,
				String: str,
			}
		} else {
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

	switch tok.Kind {
	case NumberToken:
		node = &Node{
			Kind:   NumNode,
			Value:  tok.Value,
			String: tok.String,
		}
		tok = tok.Next
		return node, tok, nil
	case IdentToken:
		node = &Node{
			Kind:   LVarNode,
			String: tok.String,
			Offset: int(tok.String[0]-'a'+1) * 8,
		}
		tok = tok.Next
		return node, tok, nil
	default:
		return node, tok, fmt.Errorf("unexpected token: %v", tok.String)
	}
}

func unary(tok *Token) (*Node, *Token, error) {
	switch tok.String {
	case "+":
		return primary(tok.Next)
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
			String: "-",
		}
		return node, tok, err
	default:
		return primary(tok)
	}
}

func (n *Node) genLocalVariable() ([]byte, error) {
	if n.Kind != LVarNode {
		return nil, fmt.Errorf("left node is not variable: %v", n.String)
	}
	return []byte(fmt.Sprintf("  mov rax, rbp\n  sub rax, %v\n  push rax\n", n.Offset)), nil
}

func (n *Node) Gen() ([]byte, error) {
	var buf []byte
	switch n.Kind {
	case NumNode:
		buf = []byte("  push ")
		buf = append(buf, n.String...)
		buf = append(buf, '\n')
		return buf, nil
	case LVarNode:
		b, err := n.genLocalVariable()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
		buf = append(buf, "  pop rax\n  mov rax, [rax]\n  push rax\n"...)
		return buf, nil
	case AssignNode:
		b, err := n.Left.genLocalVariable()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)

		b, err = n.Right.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
		buf = append(buf, "  pop rdi\n  pop rax\n  mov [rax], rdi\n  push rdi\n"...)
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
	case EqNode:
		buf = append(buf, "  cmp rax, rdi\n  sete al\n  movzb rax, al\n"...)
	case NeNode:
		buf = append(buf, "  cmp rax, rdi\n  setne al\n  movzb rax, al\n"...)
	case LtNode:
		buf = append(buf, "  cmp rax, rdi\n  setl al\n  movzb rax, al\n"...)
	case LeNode:
		buf = append(buf, "  cmp rax, rdi\n  setle al\n  movzb rax, al\n"...)
	}

	buf = append(buf, "  push rax\n"...)
	return buf, nil
}
