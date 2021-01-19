package ast

import (
	"fmt"
	"sync/atomic"
)

const (
	AddNode = iota
	SubNode
	MulNode
	DivNode
	ModNode
	NumNode
	EqNode
	NeNode
	LtNode
	LeNode
	AssignNode
	LVarNode
	ReturnNode
	IfNode
	ElseNode
	WhileNode
	ForNode
	BlockNode
)

type NKind int

type Node struct {
	Kind   NKind
	Left   *Node
	Right  *Node
	Nodes  []*Node
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
	if tok.String == "{" {
		tok = tok.Next
		n := &Node{
			Kind: BlockNode,
		}
		for tok.String != "}" {
			node, tok, err = stmt(tok)
			if err != nil {
				return node, tok, err
			}
			n.Nodes = append(n.Nodes, node)
		}
		tok = tok.Next
		return n, tok, nil
	}
	switch tok.Kind {
	case ReturnToken:
		node, tok, err = expr(tok.Next)
		if err != nil {
			return node, tok, err
		}
		node = &Node{
			Kind: ReturnNode,
			Left: node,
		}
	case IfToken:
		var (
			exprNode, stmtNode, elseStmtNode *Node
		)
		tok = tok.Next
		if tok.String != "(" {
			return node, tok, fmt.Errorf("unexpected token: %v:%v %v", tok.Row, tok.Col, tok.String)
		}
		exprNode, tok, err = expr(tok.Next)
		if err != nil {
			return node, tok, err
		}
		if tok.String != ")" {
			return node, tok, fmt.Errorf("unexpected token: %v:%v %v", tok.Row, tok.Col, tok.String)
		}
		stmtNode, tok, err = stmt(tok.Next)
		if err != nil {
			return node, tok, err
		}
		if tok.Kind == ElseToken {
			elseStmtNode, tok, err = stmt(tok.Next)
			stmtNode = &Node{
				Kind:   ElseNode,
				Left:   stmtNode,
				Right:  elseStmtNode,
				String: "else",
			}
		}
		node = &Node{
			Kind:   IfNode,
			Left:   exprNode,
			Right:  stmtNode,
			String: "if",
		}
		return node, tok, nil
	case WhileToken:
		var (
			exprNode, stmtNode *Node
		)
		tok = tok.Next
		if tok.String != "(" {
			return node, tok, fmt.Errorf("unexpected token: %v:%v %v", tok.Row, tok.Col, tok.String)
		}
		exprNode, tok, err = expr(tok.Next)
		if err != nil {
			return node, tok, err
		}
		if tok.String != ")" {
			return node, tok, fmt.Errorf("unexpected token: %v:%v %v", tok.Row, tok.Col, tok.String)
		}
		stmtNode, tok, err = stmt(tok.Next)
		if err != nil {
			return node, tok, err
		}
		node = &Node{
			Kind:   WhileNode,
			Left:   exprNode,
			Right:  stmtNode,
			String: "while",
		}
		return node, tok, nil
	default:
		node, tok, err = expr(tok)
		if err != nil {
			return node, tok, err
		}
	}
	if tok.String != ";" && tok.String != "\n" && tok.Kind != EOFToken {
		return node, tok, fmt.Errorf("unexpected token: %v:%v %v", tok.Row, tok.Col, tok.String)
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
		if str == "*" || str == "/" || str == "%" {
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
			case "%":
				k = ModNode
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
			Offset: tok.Offset,
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

var labelNum int32

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
	case ReturnNode:
		b, err := n.Left.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
		buf = append(buf, "  pop rax\n  mov rsp, rbp\n  pop rbp\n  ret\n"...)
		return buf, nil
	case IfNode:
		b, err := n.Left.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
		ln := atomic.AddInt32(&labelNum, 1)
		buf = append(buf, fmt.Sprintf("  pop rax\n  cmp rax, 0\n  je .L%v\n", ln)...)
		if n.Right.Kind == ElseNode {
			b, err = n.Right.Left.Gen()
			if err != nil {
				return nil, err
			}
			buf = append(buf, b...)
			ln2 := ln
			ln = atomic.AddInt32(&labelNum, 1)
			buf = append(buf, fmt.Sprintf("  jmp .L%v\n.L%v:\n", ln, ln2)...)
			b, err = n.Right.Right.Gen()
			if err != nil {
				return nil, err
			}
			buf = append(buf, b...)
		} else {
			b, err = n.Right.Gen()
			if err != nil {
				return nil, err
			}
			buf = append(buf, b...)
		}
		buf = append(buf, fmt.Sprintf(".L%v:\n", ln)...)
		return buf, nil
	case WhileNode:
		begin := atomic.AddInt32(&labelNum, 1)
		buf = append(buf, fmt.Sprintf(".L%v:\n", begin)...)
		b, err := n.Left.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
		end := atomic.AddInt32(&labelNum, 1)
		buf = append(buf, fmt.Sprintf("  pop rax\n  cmp rax, 0\n  je .L%v\n", end)...)
		b, err = n.Right.Gen()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b...)
		buf = append(buf, fmt.Sprintf("  jmp .L%v\n.L%v:", begin, end)...)
		return buf, err
	case BlockNode:
		for _, node := range n.Nodes {
			b, err := node.Gen()
			if err != nil {
				return nil, err
			}
			buf = append(buf, b...)
		}
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
	case ModNode:
		buf = append(buf, "  cqo\n  idiv rax, rdi\n  push rdx\n  pop rax\n"...)
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
