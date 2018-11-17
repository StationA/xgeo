package lang

//go:generate peg lang/grammar.peg

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulefile
	rulesection
	ruleblock
	ruleif_block
	rulestmt
	ruleemit
	ruleassignment
	ruleexpr
	ruleprimary
	ruleor_expr
	ruleand_expr
	rulecompare_expr
	ruleadd_expr
	rulemult_expr
	rulecompare_op
	rulefunc_call
	ruleref
	ruleglobal_ref
	rulevariable_ref
	ruleident
	ruleliteral
	rulefloat
	ruleint
	rulebool
	rulestring
	rulemap
	rulekv
	rulews
	rulewsn
)

var rul3s = [...]string{
	"Unknown",
	"file",
	"section",
	"block",
	"if_block",
	"stmt",
	"emit",
	"assignment",
	"expr",
	"primary",
	"or_expr",
	"and_expr",
	"compare_expr",
	"add_expr",
	"mult_expr",
	"compare_op",
	"func_call",
	"ref",
	"global_ref",
	"variable_ref",
	"ident",
	"literal",
	"float",
	"int",
	"bool",
	"string",
	"map",
	"kv",
	"ws",
	"wsn",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(w io.Writer, pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Fprintf(w, "%v %v\n", rule, quote)
			} else {
				fmt.Fprintf(w, "\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(w io.Writer, buffer string) {
	node.print(w, false, buffer)
}

func (node *node32) PrettyPrint(w io.Writer, buffer string) {
	node.print(w, true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(os.Stdout, buffer)
}

func (t *tokens32) WriteSyntaxTree(w io.Writer, buffer string) {
	t.AST().Print(w, buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(os.Stdout, buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type XGeoParser struct {
	Buffer string
	buffer []rune
	rules  [30]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *XGeoParser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *XGeoParser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *XGeoParser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *XGeoParser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *XGeoParser) WriteSyntaxTree(w io.Writer) {
	p.tokens32.WriteSyntaxTree(w, p.Buffer)
}

func (p *XGeoParser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 file <- <(wsn section* !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rulewsn]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3 := position, tokenIndex
					if !_rules[rulesection]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex = position3, tokenIndex3
				}
				{
					position4, tokenIndex4 := position, tokenIndex
					if !matchDot() {
						goto l4
					}
					goto l0
				l4:
					position, tokenIndex = position4, tokenIndex4
				}
				add(rulefile, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 section <- <((block / stmt) wsn)> */
		func() bool {
			position5, tokenIndex5 := position, tokenIndex
			{
				position6 := position
				{
					position7, tokenIndex7 := position, tokenIndex
					if !_rules[ruleblock]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex = position7, tokenIndex7
					if !_rules[rulestmt]() {
						goto l5
					}
				}
			l7:
				if !_rules[rulewsn]() {
					goto l5
				}
				add(rulesection, position6)
			}
			return true
		l5:
			position, tokenIndex = position5, tokenIndex5
			return false
		},
		/* 2 block <- <if_block> */
		func() bool {
			position9, tokenIndex9 := position, tokenIndex
			{
				position10 := position
				if !_rules[ruleif_block]() {
					goto l9
				}
				add(ruleblock, position10)
			}
			return true
		l9:
			position, tokenIndex = position9, tokenIndex9
			return false
		},
		/* 3 if_block <- <('i' 'f' ws '(' ws expr ws ')' ws '{' wsn section* wsn '}')> */
		func() bool {
			position11, tokenIndex11 := position, tokenIndex
			{
				position12 := position
				if buffer[position] != rune('i') {
					goto l11
				}
				position++
				if buffer[position] != rune('f') {
					goto l11
				}
				position++
				if !_rules[rulews]() {
					goto l11
				}
				if buffer[position] != rune('(') {
					goto l11
				}
				position++
				if !_rules[rulews]() {
					goto l11
				}
				if !_rules[ruleexpr]() {
					goto l11
				}
				if !_rules[rulews]() {
					goto l11
				}
				if buffer[position] != rune(')') {
					goto l11
				}
				position++
				if !_rules[rulews]() {
					goto l11
				}
				if buffer[position] != rune('{') {
					goto l11
				}
				position++
				if !_rules[rulewsn]() {
					goto l11
				}
			l13:
				{
					position14, tokenIndex14 := position, tokenIndex
					if !_rules[rulesection]() {
						goto l14
					}
					goto l13
				l14:
					position, tokenIndex = position14, tokenIndex14
				}
				if !_rules[rulewsn]() {
					goto l11
				}
				if buffer[position] != rune('}') {
					goto l11
				}
				position++
				add(ruleif_block, position12)
			}
			return true
		l11:
			position, tokenIndex = position11, tokenIndex11
			return false
		},
		/* 4 stmt <- <(emit / assignment)> */
		func() bool {
			position15, tokenIndex15 := position, tokenIndex
			{
				position16 := position
				{
					position17, tokenIndex17 := position, tokenIndex
					if !_rules[ruleemit]() {
						goto l18
					}
					goto l17
				l18:
					position, tokenIndex = position17, tokenIndex17
					if !_rules[ruleassignment]() {
						goto l15
					}
				}
			l17:
				add(rulestmt, position16)
			}
			return true
		l15:
			position, tokenIndex = position15, tokenIndex15
			return false
		},
		/* 5 emit <- <('e' 'm' 'i' 't' ws expr)> */
		func() bool {
			position19, tokenIndex19 := position, tokenIndex
			{
				position20 := position
				if buffer[position] != rune('e') {
					goto l19
				}
				position++
				if buffer[position] != rune('m') {
					goto l19
				}
				position++
				if buffer[position] != rune('i') {
					goto l19
				}
				position++
				if buffer[position] != rune('t') {
					goto l19
				}
				position++
				if !_rules[rulews]() {
					goto l19
				}
				if !_rules[ruleexpr]() {
					goto l19
				}
				add(ruleemit, position20)
			}
			return true
		l19:
			position, tokenIndex = position19, tokenIndex19
			return false
		},
		/* 6 assignment <- <(ref ws '=' ws expr)> */
		func() bool {
			position21, tokenIndex21 := position, tokenIndex
			{
				position22 := position
				if !_rules[ruleref]() {
					goto l21
				}
				if !_rules[rulews]() {
					goto l21
				}
				if buffer[position] != rune('=') {
					goto l21
				}
				position++
				if !_rules[rulews]() {
					goto l21
				}
				if !_rules[ruleexpr]() {
					goto l21
				}
				add(ruleassignment, position22)
			}
			return true
		l21:
			position, tokenIndex = position21, tokenIndex21
			return false
		},
		/* 7 expr <- <(or_expr / primary)> */
		func() bool {
			position23, tokenIndex23 := position, tokenIndex
			{
				position24 := position
				{
					position25, tokenIndex25 := position, tokenIndex
					if !_rules[ruleor_expr]() {
						goto l26
					}
					goto l25
				l26:
					position, tokenIndex = position25, tokenIndex25
					if !_rules[ruleprimary]() {
						goto l23
					}
				}
			l25:
				add(ruleexpr, position24)
			}
			return true
		l23:
			position, tokenIndex = position23, tokenIndex23
			return false
		},
		/* 8 primary <- <(('(' ws expr ws ')') / func_call / literal / ref)> */
		func() bool {
			position27, tokenIndex27 := position, tokenIndex
			{
				position28 := position
				{
					position29, tokenIndex29 := position, tokenIndex
					if buffer[position] != rune('(') {
						goto l30
					}
					position++
					if !_rules[rulews]() {
						goto l30
					}
					if !_rules[ruleexpr]() {
						goto l30
					}
					if !_rules[rulews]() {
						goto l30
					}
					if buffer[position] != rune(')') {
						goto l30
					}
					position++
					goto l29
				l30:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[rulefunc_call]() {
						goto l31
					}
					goto l29
				l31:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleliteral]() {
						goto l32
					}
					goto l29
				l32:
					position, tokenIndex = position29, tokenIndex29
					if !_rules[ruleref]() {
						goto l27
					}
				}
			l29:
				add(ruleprimary, position28)
			}
			return true
		l27:
			position, tokenIndex = position27, tokenIndex27
			return false
		},
		/* 9 or_expr <- <(and_expr (ws ('|' '|') ws and_expr)*)> */
		func() bool {
			position33, tokenIndex33 := position, tokenIndex
			{
				position34 := position
				if !_rules[ruleand_expr]() {
					goto l33
				}
			l35:
				{
					position36, tokenIndex36 := position, tokenIndex
					if !_rules[rulews]() {
						goto l36
					}
					if buffer[position] != rune('|') {
						goto l36
					}
					position++
					if buffer[position] != rune('|') {
						goto l36
					}
					position++
					if !_rules[rulews]() {
						goto l36
					}
					if !_rules[ruleand_expr]() {
						goto l36
					}
					goto l35
				l36:
					position, tokenIndex = position36, tokenIndex36
				}
				add(ruleor_expr, position34)
			}
			return true
		l33:
			position, tokenIndex = position33, tokenIndex33
			return false
		},
		/* 10 and_expr <- <(compare_expr (ws ('&' '&') ws compare_expr)*)> */
		func() bool {
			position37, tokenIndex37 := position, tokenIndex
			{
				position38 := position
				if !_rules[rulecompare_expr]() {
					goto l37
				}
			l39:
				{
					position40, tokenIndex40 := position, tokenIndex
					if !_rules[rulews]() {
						goto l40
					}
					if buffer[position] != rune('&') {
						goto l40
					}
					position++
					if buffer[position] != rune('&') {
						goto l40
					}
					position++
					if !_rules[rulews]() {
						goto l40
					}
					if !_rules[rulecompare_expr]() {
						goto l40
					}
					goto l39
				l40:
					position, tokenIndex = position40, tokenIndex40
				}
				add(ruleand_expr, position38)
			}
			return true
		l37:
			position, tokenIndex = position37, tokenIndex37
			return false
		},
		/* 11 compare_expr <- <(add_expr (ws compare_op ws add_expr)*)> */
		func() bool {
			position41, tokenIndex41 := position, tokenIndex
			{
				position42 := position
				if !_rules[ruleadd_expr]() {
					goto l41
				}
			l43:
				{
					position44, tokenIndex44 := position, tokenIndex
					if !_rules[rulews]() {
						goto l44
					}
					if !_rules[rulecompare_op]() {
						goto l44
					}
					if !_rules[rulews]() {
						goto l44
					}
					if !_rules[ruleadd_expr]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex = position44, tokenIndex44
				}
				add(rulecompare_expr, position42)
			}
			return true
		l41:
			position, tokenIndex = position41, tokenIndex41
			return false
		},
		/* 12 add_expr <- <(mult_expr (ws ('+' / '-') ws mult_expr)*)> */
		func() bool {
			position45, tokenIndex45 := position, tokenIndex
			{
				position46 := position
				if !_rules[rulemult_expr]() {
					goto l45
				}
			l47:
				{
					position48, tokenIndex48 := position, tokenIndex
					if !_rules[rulews]() {
						goto l48
					}
					{
						position49, tokenIndex49 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l50
						}
						position++
						goto l49
					l50:
						position, tokenIndex = position49, tokenIndex49
						if buffer[position] != rune('-') {
							goto l48
						}
						position++
					}
				l49:
					if !_rules[rulews]() {
						goto l48
					}
					if !_rules[rulemult_expr]() {
						goto l48
					}
					goto l47
				l48:
					position, tokenIndex = position48, tokenIndex48
				}
				add(ruleadd_expr, position46)
			}
			return true
		l45:
			position, tokenIndex = position45, tokenIndex45
			return false
		},
		/* 13 mult_expr <- <(primary (ws ('*' / '/') ws primary)*)> */
		func() bool {
			position51, tokenIndex51 := position, tokenIndex
			{
				position52 := position
				if !_rules[ruleprimary]() {
					goto l51
				}
			l53:
				{
					position54, tokenIndex54 := position, tokenIndex
					if !_rules[rulews]() {
						goto l54
					}
					{
						position55, tokenIndex55 := position, tokenIndex
						if buffer[position] != rune('*') {
							goto l56
						}
						position++
						goto l55
					l56:
						position, tokenIndex = position55, tokenIndex55
						if buffer[position] != rune('/') {
							goto l54
						}
						position++
					}
				l55:
					if !_rules[rulews]() {
						goto l54
					}
					if !_rules[ruleprimary]() {
						goto l54
					}
					goto l53
				l54:
					position, tokenIndex = position54, tokenIndex54
				}
				add(rulemult_expr, position52)
			}
			return true
		l51:
			position, tokenIndex = position51, tokenIndex51
			return false
		},
		/* 14 compare_op <- <(('=' '=') / ('!' '=') / ('>' '=') / ('<' '=') / '>' / '<')> */
		func() bool {
			position57, tokenIndex57 := position, tokenIndex
			{
				position58 := position
				{
					position59, tokenIndex59 := position, tokenIndex
					if buffer[position] != rune('=') {
						goto l60
					}
					position++
					if buffer[position] != rune('=') {
						goto l60
					}
					position++
					goto l59
				l60:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('!') {
						goto l61
					}
					position++
					if buffer[position] != rune('=') {
						goto l61
					}
					position++
					goto l59
				l61:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('>') {
						goto l62
					}
					position++
					if buffer[position] != rune('=') {
						goto l62
					}
					position++
					goto l59
				l62:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('<') {
						goto l63
					}
					position++
					if buffer[position] != rune('=') {
						goto l63
					}
					position++
					goto l59
				l63:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('>') {
						goto l64
					}
					position++
					goto l59
				l64:
					position, tokenIndex = position59, tokenIndex59
					if buffer[position] != rune('<') {
						goto l57
					}
					position++
				}
			l59:
				add(rulecompare_op, position58)
			}
			return true
		l57:
			position, tokenIndex = position57, tokenIndex57
			return false
		},
		/* 15 func_call <- <(ident '(' wsn (expr (wsn ',' wsn expr)*)? wsn ')')> */
		func() bool {
			position65, tokenIndex65 := position, tokenIndex
			{
				position66 := position
				if !_rules[ruleident]() {
					goto l65
				}
				if buffer[position] != rune('(') {
					goto l65
				}
				position++
				if !_rules[rulewsn]() {
					goto l65
				}
				{
					position67, tokenIndex67 := position, tokenIndex
					if !_rules[ruleexpr]() {
						goto l67
					}
				l69:
					{
						position70, tokenIndex70 := position, tokenIndex
						if !_rules[rulewsn]() {
							goto l70
						}
						if buffer[position] != rune(',') {
							goto l70
						}
						position++
						if !_rules[rulewsn]() {
							goto l70
						}
						if !_rules[ruleexpr]() {
							goto l70
						}
						goto l69
					l70:
						position, tokenIndex = position70, tokenIndex70
					}
					goto l68
				l67:
					position, tokenIndex = position67, tokenIndex67
				}
			l68:
				if !_rules[rulewsn]() {
					goto l65
				}
				if buffer[position] != rune(')') {
					goto l65
				}
				position++
				add(rulefunc_call, position66)
			}
			return true
		l65:
			position, tokenIndex = position65, tokenIndex65
			return false
		},
		/* 16 ref <- <((global_ref / variable_ref) ('.' variable_ref)*)> */
		func() bool {
			position71, tokenIndex71 := position, tokenIndex
			{
				position72 := position
				{
					position73, tokenIndex73 := position, tokenIndex
					if !_rules[ruleglobal_ref]() {
						goto l74
					}
					goto l73
				l74:
					position, tokenIndex = position73, tokenIndex73
					if !_rules[rulevariable_ref]() {
						goto l71
					}
				}
			l73:
			l75:
				{
					position76, tokenIndex76 := position, tokenIndex
					if buffer[position] != rune('.') {
						goto l76
					}
					position++
					if !_rules[rulevariable_ref]() {
						goto l76
					}
					goto l75
				l76:
					position, tokenIndex = position76, tokenIndex76
				}
				add(ruleref, position72)
			}
			return true
		l71:
			position, tokenIndex = position71, tokenIndex71
			return false
		},
		/* 17 global_ref <- <('@' ident?)> */
		func() bool {
			position77, tokenIndex77 := position, tokenIndex
			{
				position78 := position
				if buffer[position] != rune('@') {
					goto l77
				}
				position++
				{
					position79, tokenIndex79 := position, tokenIndex
					if !_rules[ruleident]() {
						goto l79
					}
					goto l80
				l79:
					position, tokenIndex = position79, tokenIndex79
				}
			l80:
				add(ruleglobal_ref, position78)
			}
			return true
		l77:
			position, tokenIndex = position77, tokenIndex77
			return false
		},
		/* 18 variable_ref <- <ident> */
		func() bool {
			position81, tokenIndex81 := position, tokenIndex
			{
				position82 := position
				if !_rules[ruleident]() {
					goto l81
				}
				add(rulevariable_ref, position82)
			}
			return true
		l81:
			position, tokenIndex = position81, tokenIndex81
			return false
		},
		/* 19 ident <- <(([A-Z] / [a-z] / '_') ([A-Z] / [a-z] / [0-9] / '_')*)> */
		func() bool {
			position83, tokenIndex83 := position, tokenIndex
			{
				position84 := position
				{
					position85, tokenIndex85 := position, tokenIndex
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex = position85, tokenIndex85
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l87
					}
					position++
					goto l85
				l87:
					position, tokenIndex = position85, tokenIndex85
					if buffer[position] != rune('_') {
						goto l83
					}
					position++
				}
			l85:
			l88:
				{
					position89, tokenIndex89 := position, tokenIndex
					{
						position90, tokenIndex90 := position, tokenIndex
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l91
						}
						position++
						goto l90
					l91:
						position, tokenIndex = position90, tokenIndex90
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l92
						}
						position++
						goto l90
					l92:
						position, tokenIndex = position90, tokenIndex90
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l93
						}
						position++
						goto l90
					l93:
						position, tokenIndex = position90, tokenIndex90
						if buffer[position] != rune('_') {
							goto l89
						}
						position++
					}
				l90:
					goto l88
				l89:
					position, tokenIndex = position89, tokenIndex89
				}
				add(ruleident, position84)
			}
			return true
		l83:
			position, tokenIndex = position83, tokenIndex83
			return false
		},
		/* 20 literal <- <(bool / float / int / string / map)> */
		func() bool {
			position94, tokenIndex94 := position, tokenIndex
			{
				position95 := position
				{
					position96, tokenIndex96 := position, tokenIndex
					if !_rules[rulebool]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex = position96, tokenIndex96
					if !_rules[rulefloat]() {
						goto l98
					}
					goto l96
				l98:
					position, tokenIndex = position96, tokenIndex96
					if !_rules[ruleint]() {
						goto l99
					}
					goto l96
				l99:
					position, tokenIndex = position96, tokenIndex96
					if !_rules[rulestring]() {
						goto l100
					}
					goto l96
				l100:
					position, tokenIndex = position96, tokenIndex96
					if !_rules[rulemap]() {
						goto l94
					}
				}
			l96:
				add(ruleliteral, position95)
			}
			return true
		l94:
			position, tokenIndex = position94, tokenIndex94
			return false
		},
		/* 21 float <- <('-'? [0-9]+ '.' [0-9]*)> */
		func() bool {
			position101, tokenIndex101 := position, tokenIndex
			{
				position102 := position
				{
					position103, tokenIndex103 := position, tokenIndex
					if buffer[position] != rune('-') {
						goto l103
					}
					position++
					goto l104
				l103:
					position, tokenIndex = position103, tokenIndex103
				}
			l104:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l101
				}
				position++
			l105:
				{
					position106, tokenIndex106 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l106
					}
					position++
					goto l105
				l106:
					position, tokenIndex = position106, tokenIndex106
				}
				if buffer[position] != rune('.') {
					goto l101
				}
				position++
			l107:
				{
					position108, tokenIndex108 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l108
					}
					position++
					goto l107
				l108:
					position, tokenIndex = position108, tokenIndex108
				}
				add(rulefloat, position102)
			}
			return true
		l101:
			position, tokenIndex = position101, tokenIndex101
			return false
		},
		/* 22 int <- <('-'? [0-9]+)> */
		func() bool {
			position109, tokenIndex109 := position, tokenIndex
			{
				position110 := position
				{
					position111, tokenIndex111 := position, tokenIndex
					if buffer[position] != rune('-') {
						goto l111
					}
					position++
					goto l112
				l111:
					position, tokenIndex = position111, tokenIndex111
				}
			l112:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l109
				}
				position++
			l113:
				{
					position114, tokenIndex114 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l114
					}
					position++
					goto l113
				l114:
					position, tokenIndex = position114, tokenIndex114
				}
				add(ruleint, position110)
			}
			return true
		l109:
			position, tokenIndex = position109, tokenIndex109
			return false
		},
		/* 23 bool <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position115, tokenIndex115 := position, tokenIndex
			{
				position116 := position
				{
					position117, tokenIndex117 := position, tokenIndex
					if buffer[position] != rune('t') {
						goto l118
					}
					position++
					if buffer[position] != rune('r') {
						goto l118
					}
					position++
					if buffer[position] != rune('u') {
						goto l118
					}
					position++
					if buffer[position] != rune('e') {
						goto l118
					}
					position++
					goto l117
				l118:
					position, tokenIndex = position117, tokenIndex117
					if buffer[position] != rune('f') {
						goto l115
					}
					position++
					if buffer[position] != rune('a') {
						goto l115
					}
					position++
					if buffer[position] != rune('l') {
						goto l115
					}
					position++
					if buffer[position] != rune('s') {
						goto l115
					}
					position++
					if buffer[position] != rune('e') {
						goto l115
					}
					position++
				}
			l117:
				add(rulebool, position116)
			}
			return true
		l115:
			position, tokenIndex = position115, tokenIndex115
			return false
		},
		/* 24 string <- <('"' (!'"' .)* '"')> */
		func() bool {
			position119, tokenIndex119 := position, tokenIndex
			{
				position120 := position
				if buffer[position] != rune('"') {
					goto l119
				}
				position++
			l121:
				{
					position122, tokenIndex122 := position, tokenIndex
					{
						position123, tokenIndex123 := position, tokenIndex
						if buffer[position] != rune('"') {
							goto l123
						}
						position++
						goto l122
					l123:
						position, tokenIndex = position123, tokenIndex123
					}
					if !matchDot() {
						goto l122
					}
					goto l121
				l122:
					position, tokenIndex = position122, tokenIndex122
				}
				if buffer[position] != rune('"') {
					goto l119
				}
				position++
				add(rulestring, position120)
			}
			return true
		l119:
			position, tokenIndex = position119, tokenIndex119
			return false
		},
		/* 25 map <- <('{' wsn (kv (',' wsn kv)*)? wsn '}')> */
		func() bool {
			position124, tokenIndex124 := position, tokenIndex
			{
				position125 := position
				if buffer[position] != rune('{') {
					goto l124
				}
				position++
				if !_rules[rulewsn]() {
					goto l124
				}
				{
					position126, tokenIndex126 := position, tokenIndex
					if !_rules[rulekv]() {
						goto l126
					}
				l128:
					{
						position129, tokenIndex129 := position, tokenIndex
						if buffer[position] != rune(',') {
							goto l129
						}
						position++
						if !_rules[rulewsn]() {
							goto l129
						}
						if !_rules[rulekv]() {
							goto l129
						}
						goto l128
					l129:
						position, tokenIndex = position129, tokenIndex129
					}
					goto l127
				l126:
					position, tokenIndex = position126, tokenIndex126
				}
			l127:
				if !_rules[rulewsn]() {
					goto l124
				}
				if buffer[position] != rune('}') {
					goto l124
				}
				position++
				add(rulemap, position125)
			}
			return true
		l124:
			position, tokenIndex = position124, tokenIndex124
			return false
		},
		/* 26 kv <- <(string ws ':' wsn expr)> */
		func() bool {
			position130, tokenIndex130 := position, tokenIndex
			{
				position131 := position
				if !_rules[rulestring]() {
					goto l130
				}
				if !_rules[rulews]() {
					goto l130
				}
				if buffer[position] != rune(':') {
					goto l130
				}
				position++
				if !_rules[rulewsn]() {
					goto l130
				}
				if !_rules[ruleexpr]() {
					goto l130
				}
				add(rulekv, position131)
			}
			return true
		l130:
			position, tokenIndex = position130, tokenIndex130
			return false
		},
		/* 27 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position133 := position
			l134:
				{
					position135, tokenIndex135 := position, tokenIndex
					{
						position136, tokenIndex136 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l137
						}
						position++
						goto l136
					l137:
						position, tokenIndex = position136, tokenIndex136
						if buffer[position] != rune('\t') {
							goto l135
						}
						position++
					}
				l136:
					goto l134
				l135:
					position, tokenIndex = position135, tokenIndex135
				}
				add(rulews, position133)
			}
			return true
		},
		/* 28 wsn <- <(' ' / '\t' / '\n')*> */
		func() bool {
			{
				position139 := position
			l140:
				{
					position141, tokenIndex141 := position, tokenIndex
					{
						position142, tokenIndex142 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l143
						}
						position++
						goto l142
					l143:
						position, tokenIndex = position142, tokenIndex142
						if buffer[position] != rune('\t') {
							goto l144
						}
						position++
						goto l142
					l144:
						position, tokenIndex = position142, tokenIndex142
						if buffer[position] != rune('\n') {
							goto l141
						}
						position++
					}
				l142:
					goto l140
				l141:
					position, tokenIndex = position141, tokenIndex141
				}
				add(rulewsn, position139)
			}
			return true
		},
	}
	p.rules = _rules
}
