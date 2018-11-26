package lang

//go:generate peg lang/grammar.peg

import (
	"fmt"
	"github.com/stationa/xgeo/vm"
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
	ruleif_cond
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
	rulefunc_call
	ruleref
	rulederef
	ruleglobal_ref
	rulevariable_ref
	ruleident
	ruleliteral
	rulefloat
	ruleint
	rulebool
	rulestring
	rulews
	rulewsn
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
)

var rul3s = [...]string{
	"Unknown",
	"file",
	"section",
	"block",
	"if_block",
	"if_cond",
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
	"func_call",
	"ref",
	"deref",
	"global_ref",
	"variable_ref",
	"ident",
	"literal",
	"float",
	"int",
	"bool",
	"string",
	"ws",
	"wsn",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
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

type XGeoCompiler struct {
	constants     []vm.Value
	code          []*vm.Code
	jumpStack     []*vm.Code
	registerCount int
	refs          map[string]int

	Buffer string
	buffer []rune
	rules  [53]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *XGeoCompiler) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *XGeoCompiler) Reset() {
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
	p   *XGeoCompiler
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

func (p *XGeoCompiler) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *XGeoCompiler) WriteSyntaxTree(w io.Writer) {
	p.tokens32.WriteSyntaxTree(w, p.Buffer)
}

func (p *XGeoCompiler) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:

			p.SetJump()

		case ruleAction1:

			p.AddCondJump()

		case ruleAction2:

			p.AddCode(vm.OpEMIT)

		case ruleAction3:

			p.AllocateRef(buffer[begin:end])

		case ruleAction4:

			p.AddStore()

		case ruleAction5:
			p.AddCode(vm.OpAND)
		case ruleAction6:
			p.AddCode(vm.OpOR)
		case ruleAction7:
			p.AddCode(vm.OpEQ)
		case ruleAction8:
			p.AddCode(vm.OpNEQ)
		case ruleAction9:
			p.AddCode(vm.OpLT)
		case ruleAction10:
			p.AddCode(vm.OpLTE)
		case ruleAction11:
			p.AddCode(vm.OpGT)
		case ruleAction12:
			p.AddCode(vm.OpGTE)
		case ruleAction13:
			p.AddCode(vm.OpADD)
		case ruleAction14:
			p.AddCode(vm.OpSUB)
		case ruleAction15:
			p.AddCode(vm.OpMUL)
		case ruleAction16:
			p.AddCode(vm.OpDIV)
		case ruleAction17:

			p.AddCode(vm.OpCALL)

		case ruleAction18:

			p.AddLoad(buffer[begin:end])

		case ruleAction19:

			p.AddConstant(ParseFloat(buffer[begin:end]))

		case ruleAction20:

			p.AddConstant(ParseInt(buffer[begin:end]))

		case ruleAction21:

			p.AddConstant(ParseBool(buffer[begin:end]))

		case ruleAction22:

			p.AddConstant(&vm.Str{buffer[begin:end]})

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *XGeoCompiler) Init() {
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
		/* 1 section <- <(block / stmt)> */
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
		/* 3 if_block <- <(if_cond '{' wsn section* wsn '}' wsn Action0)> */
		func() bool {
			position11, tokenIndex11 := position, tokenIndex
			{
				position12 := position
				if !_rules[ruleif_cond]() {
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
				if !_rules[rulewsn]() {
					goto l11
				}
				if !_rules[ruleAction0]() {
					goto l11
				}
				add(ruleif_block, position12)
			}
			return true
		l11:
			position, tokenIndex = position11, tokenIndex11
			return false
		},
		/* 4 if_cond <- <('i' 'f' ws '(' ws expr ws ')' wsn Action1)> */
		func() bool {
			position15, tokenIndex15 := position, tokenIndex
			{
				position16 := position
				if buffer[position] != rune('i') {
					goto l15
				}
				position++
				if buffer[position] != rune('f') {
					goto l15
				}
				position++
				if !_rules[rulews]() {
					goto l15
				}
				if buffer[position] != rune('(') {
					goto l15
				}
				position++
				if !_rules[rulews]() {
					goto l15
				}
				if !_rules[ruleexpr]() {
					goto l15
				}
				if !_rules[rulews]() {
					goto l15
				}
				if buffer[position] != rune(')') {
					goto l15
				}
				position++
				if !_rules[rulewsn]() {
					goto l15
				}
				if !_rules[ruleAction1]() {
					goto l15
				}
				add(ruleif_cond, position16)
			}
			return true
		l15:
			position, tokenIndex = position15, tokenIndex15
			return false
		},
		/* 5 stmt <- <((emit / assignment) wsn)> */
		func() bool {
			position17, tokenIndex17 := position, tokenIndex
			{
				position18 := position
				{
					position19, tokenIndex19 := position, tokenIndex
					if !_rules[ruleemit]() {
						goto l20
					}
					goto l19
				l20:
					position, tokenIndex = position19, tokenIndex19
					if !_rules[ruleassignment]() {
						goto l17
					}
				}
			l19:
				if !_rules[rulewsn]() {
					goto l17
				}
				add(rulestmt, position18)
			}
			return true
		l17:
			position, tokenIndex = position17, tokenIndex17
			return false
		},
		/* 6 emit <- <('e' 'm' 'i' 't' ws expr Action2)> */
		func() bool {
			position21, tokenIndex21 := position, tokenIndex
			{
				position22 := position
				if buffer[position] != rune('e') {
					goto l21
				}
				position++
				if buffer[position] != rune('m') {
					goto l21
				}
				position++
				if buffer[position] != rune('i') {
					goto l21
				}
				position++
				if buffer[position] != rune('t') {
					goto l21
				}
				position++
				if !_rules[rulews]() {
					goto l21
				}
				if !_rules[ruleexpr]() {
					goto l21
				}
				if !_rules[ruleAction2]() {
					goto l21
				}
				add(ruleemit, position22)
			}
			return true
		l21:
			position, tokenIndex = position21, tokenIndex21
			return false
		},
		/* 7 assignment <- <(<ref> Action3 ws '=' ws expr Action4)> */
		func() bool {
			position23, tokenIndex23 := position, tokenIndex
			{
				position24 := position
				{
					position25 := position
					if !_rules[ruleref]() {
						goto l23
					}
					add(rulePegText, position25)
				}
				if !_rules[ruleAction3]() {
					goto l23
				}
				if !_rules[rulews]() {
					goto l23
				}
				if buffer[position] != rune('=') {
					goto l23
				}
				position++
				if !_rules[rulews]() {
					goto l23
				}
				if !_rules[ruleexpr]() {
					goto l23
				}
				if !_rules[ruleAction4]() {
					goto l23
				}
				add(ruleassignment, position24)
			}
			return true
		l23:
			position, tokenIndex = position23, tokenIndex23
			return false
		},
		/* 8 expr <- <(or_expr / primary)> */
		func() bool {
			position26, tokenIndex26 := position, tokenIndex
			{
				position27 := position
				{
					position28, tokenIndex28 := position, tokenIndex
					if !_rules[ruleor_expr]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex = position28, tokenIndex28
					if !_rules[ruleprimary]() {
						goto l26
					}
				}
			l28:
				add(ruleexpr, position27)
			}
			return true
		l26:
			position, tokenIndex = position26, tokenIndex26
			return false
		},
		/* 9 primary <- <(('(' ws expr ws ')') / func_call / literal / deref)> */
		func() bool {
			position30, tokenIndex30 := position, tokenIndex
			{
				position31 := position
				{
					position32, tokenIndex32 := position, tokenIndex
					if buffer[position] != rune('(') {
						goto l33
					}
					position++
					if !_rules[rulews]() {
						goto l33
					}
					if !_rules[ruleexpr]() {
						goto l33
					}
					if !_rules[rulews]() {
						goto l33
					}
					if buffer[position] != rune(')') {
						goto l33
					}
					position++
					goto l32
				l33:
					position, tokenIndex = position32, tokenIndex32
					if !_rules[rulefunc_call]() {
						goto l34
					}
					goto l32
				l34:
					position, tokenIndex = position32, tokenIndex32
					if !_rules[ruleliteral]() {
						goto l35
					}
					goto l32
				l35:
					position, tokenIndex = position32, tokenIndex32
					if !_rules[rulederef]() {
						goto l30
					}
				}
			l32:
				add(ruleprimary, position31)
			}
			return true
		l30:
			position, tokenIndex = position30, tokenIndex30
			return false
		},
		/* 10 or_expr <- <(and_expr (ws ('|' '|') ws expr Action5)*)> */
		func() bool {
			position36, tokenIndex36 := position, tokenIndex
			{
				position37 := position
				if !_rules[ruleand_expr]() {
					goto l36
				}
			l38:
				{
					position39, tokenIndex39 := position, tokenIndex
					if !_rules[rulews]() {
						goto l39
					}
					if buffer[position] != rune('|') {
						goto l39
					}
					position++
					if buffer[position] != rune('|') {
						goto l39
					}
					position++
					if !_rules[rulews]() {
						goto l39
					}
					if !_rules[ruleexpr]() {
						goto l39
					}
					if !_rules[ruleAction5]() {
						goto l39
					}
					goto l38
				l39:
					position, tokenIndex = position39, tokenIndex39
				}
				add(ruleor_expr, position37)
			}
			return true
		l36:
			position, tokenIndex = position36, tokenIndex36
			return false
		},
		/* 11 and_expr <- <(compare_expr (ws ('&' '&') ws expr Action6)*)> */
		func() bool {
			position40, tokenIndex40 := position, tokenIndex
			{
				position41 := position
				if !_rules[rulecompare_expr]() {
					goto l40
				}
			l42:
				{
					position43, tokenIndex43 := position, tokenIndex
					if !_rules[rulews]() {
						goto l43
					}
					if buffer[position] != rune('&') {
						goto l43
					}
					position++
					if buffer[position] != rune('&') {
						goto l43
					}
					position++
					if !_rules[rulews]() {
						goto l43
					}
					if !_rules[ruleexpr]() {
						goto l43
					}
					if !_rules[ruleAction6]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex = position43, tokenIndex43
				}
				add(ruleand_expr, position41)
			}
			return true
		l40:
			position, tokenIndex = position40, tokenIndex40
			return false
		},
		/* 12 compare_expr <- <(add_expr (ws (('=' '=' ws expr Action7) / ('!' '=' ws expr Action8) / ('<' ws expr Action9) / ('<' '=' ws expr Action10) / ('>' ws expr Action11) / ('>' '=' ws expr Action12)))*)> */
		func() bool {
			position44, tokenIndex44 := position, tokenIndex
			{
				position45 := position
				if !_rules[ruleadd_expr]() {
					goto l44
				}
			l46:
				{
					position47, tokenIndex47 := position, tokenIndex
					if !_rules[rulews]() {
						goto l47
					}
					{
						position48, tokenIndex48 := position, tokenIndex
						if buffer[position] != rune('=') {
							goto l49
						}
						position++
						if buffer[position] != rune('=') {
							goto l49
						}
						position++
						if !_rules[rulews]() {
							goto l49
						}
						if !_rules[ruleexpr]() {
							goto l49
						}
						if !_rules[ruleAction7]() {
							goto l49
						}
						goto l48
					l49:
						position, tokenIndex = position48, tokenIndex48
						if buffer[position] != rune('!') {
							goto l50
						}
						position++
						if buffer[position] != rune('=') {
							goto l50
						}
						position++
						if !_rules[rulews]() {
							goto l50
						}
						if !_rules[ruleexpr]() {
							goto l50
						}
						if !_rules[ruleAction8]() {
							goto l50
						}
						goto l48
					l50:
						position, tokenIndex = position48, tokenIndex48
						if buffer[position] != rune('<') {
							goto l51
						}
						position++
						if !_rules[rulews]() {
							goto l51
						}
						if !_rules[ruleexpr]() {
							goto l51
						}
						if !_rules[ruleAction9]() {
							goto l51
						}
						goto l48
					l51:
						position, tokenIndex = position48, tokenIndex48
						if buffer[position] != rune('<') {
							goto l52
						}
						position++
						if buffer[position] != rune('=') {
							goto l52
						}
						position++
						if !_rules[rulews]() {
							goto l52
						}
						if !_rules[ruleexpr]() {
							goto l52
						}
						if !_rules[ruleAction10]() {
							goto l52
						}
						goto l48
					l52:
						position, tokenIndex = position48, tokenIndex48
						if buffer[position] != rune('>') {
							goto l53
						}
						position++
						if !_rules[rulews]() {
							goto l53
						}
						if !_rules[ruleexpr]() {
							goto l53
						}
						if !_rules[ruleAction11]() {
							goto l53
						}
						goto l48
					l53:
						position, tokenIndex = position48, tokenIndex48
						if buffer[position] != rune('>') {
							goto l47
						}
						position++
						if buffer[position] != rune('=') {
							goto l47
						}
						position++
						if !_rules[rulews]() {
							goto l47
						}
						if !_rules[ruleexpr]() {
							goto l47
						}
						if !_rules[ruleAction12]() {
							goto l47
						}
					}
				l48:
					goto l46
				l47:
					position, tokenIndex = position47, tokenIndex47
				}
				add(rulecompare_expr, position45)
			}
			return true
		l44:
			position, tokenIndex = position44, tokenIndex44
			return false
		},
		/* 13 add_expr <- <(mult_expr (ws (('+' ws expr Action13) / ('-' ws expr Action14)))*)> */
		func() bool {
			position54, tokenIndex54 := position, tokenIndex
			{
				position55 := position
				if !_rules[rulemult_expr]() {
					goto l54
				}
			l56:
				{
					position57, tokenIndex57 := position, tokenIndex
					if !_rules[rulews]() {
						goto l57
					}
					{
						position58, tokenIndex58 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l59
						}
						position++
						if !_rules[rulews]() {
							goto l59
						}
						if !_rules[ruleexpr]() {
							goto l59
						}
						if !_rules[ruleAction13]() {
							goto l59
						}
						goto l58
					l59:
						position, tokenIndex = position58, tokenIndex58
						if buffer[position] != rune('-') {
							goto l57
						}
						position++
						if !_rules[rulews]() {
							goto l57
						}
						if !_rules[ruleexpr]() {
							goto l57
						}
						if !_rules[ruleAction14]() {
							goto l57
						}
					}
				l58:
					goto l56
				l57:
					position, tokenIndex = position57, tokenIndex57
				}
				add(ruleadd_expr, position55)
			}
			return true
		l54:
			position, tokenIndex = position54, tokenIndex54
			return false
		},
		/* 14 mult_expr <- <(primary (ws (('*' ws expr Action15) / ('/' ws expr Action16)))*)> */
		func() bool {
			position60, tokenIndex60 := position, tokenIndex
			{
				position61 := position
				if !_rules[ruleprimary]() {
					goto l60
				}
			l62:
				{
					position63, tokenIndex63 := position, tokenIndex
					if !_rules[rulews]() {
						goto l63
					}
					{
						position64, tokenIndex64 := position, tokenIndex
						if buffer[position] != rune('*') {
							goto l65
						}
						position++
						if !_rules[rulews]() {
							goto l65
						}
						if !_rules[ruleexpr]() {
							goto l65
						}
						if !_rules[ruleAction15]() {
							goto l65
						}
						goto l64
					l65:
						position, tokenIndex = position64, tokenIndex64
						if buffer[position] != rune('/') {
							goto l63
						}
						position++
						if !_rules[rulews]() {
							goto l63
						}
						if !_rules[ruleexpr]() {
							goto l63
						}
						if !_rules[ruleAction16]() {
							goto l63
						}
					}
				l64:
					goto l62
				l63:
					position, tokenIndex = position63, tokenIndex63
				}
				add(rulemult_expr, position61)
			}
			return true
		l60:
			position, tokenIndex = position60, tokenIndex60
			return false
		},
		/* 15 func_call <- <(deref '(' wsn (expr (wsn ',' wsn expr)*)? wsn ')' Action17)> */
		func() bool {
			position66, tokenIndex66 := position, tokenIndex
			{
				position67 := position
				if !_rules[rulederef]() {
					goto l66
				}
				if buffer[position] != rune('(') {
					goto l66
				}
				position++
				if !_rules[rulewsn]() {
					goto l66
				}
				{
					position68, tokenIndex68 := position, tokenIndex
					if !_rules[ruleexpr]() {
						goto l68
					}
				l70:
					{
						position71, tokenIndex71 := position, tokenIndex
						if !_rules[rulewsn]() {
							goto l71
						}
						if buffer[position] != rune(',') {
							goto l71
						}
						position++
						if !_rules[rulewsn]() {
							goto l71
						}
						if !_rules[ruleexpr]() {
							goto l71
						}
						goto l70
					l71:
						position, tokenIndex = position71, tokenIndex71
					}
					goto l69
				l68:
					position, tokenIndex = position68, tokenIndex68
				}
			l69:
				if !_rules[rulewsn]() {
					goto l66
				}
				if buffer[position] != rune(')') {
					goto l66
				}
				position++
				if !_rules[ruleAction17]() {
					goto l66
				}
				add(rulefunc_call, position67)
			}
			return true
		l66:
			position, tokenIndex = position66, tokenIndex66
			return false
		},
		/* 16 ref <- <(global_ref / variable_ref)> */
		func() bool {
			position72, tokenIndex72 := position, tokenIndex
			{
				position73 := position
				{
					position74, tokenIndex74 := position, tokenIndex
					if !_rules[ruleglobal_ref]() {
						goto l75
					}
					goto l74
				l75:
					position, tokenIndex = position74, tokenIndex74
					if !_rules[rulevariable_ref]() {
						goto l72
					}
				}
			l74:
				add(ruleref, position73)
			}
			return true
		l72:
			position, tokenIndex = position72, tokenIndex72
			return false
		},
		/* 17 deref <- <(<ref> Action18)> */
		func() bool {
			position76, tokenIndex76 := position, tokenIndex
			{
				position77 := position
				{
					position78 := position
					if !_rules[ruleref]() {
						goto l76
					}
					add(rulePegText, position78)
				}
				if !_rules[ruleAction18]() {
					goto l76
				}
				add(rulederef, position77)
			}
			return true
		l76:
			position, tokenIndex = position76, tokenIndex76
			return false
		},
		/* 18 global_ref <- <('@' (ident ('.' ident)*)?)> */
		func() bool {
			position79, tokenIndex79 := position, tokenIndex
			{
				position80 := position
				if buffer[position] != rune('@') {
					goto l79
				}
				position++
				{
					position81, tokenIndex81 := position, tokenIndex
					if !_rules[ruleident]() {
						goto l81
					}
				l83:
					{
						position84, tokenIndex84 := position, tokenIndex
						if buffer[position] != rune('.') {
							goto l84
						}
						position++
						if !_rules[ruleident]() {
							goto l84
						}
						goto l83
					l84:
						position, tokenIndex = position84, tokenIndex84
					}
					goto l82
				l81:
					position, tokenIndex = position81, tokenIndex81
				}
			l82:
				add(ruleglobal_ref, position80)
			}
			return true
		l79:
			position, tokenIndex = position79, tokenIndex79
			return false
		},
		/* 19 variable_ref <- <ident> */
		func() bool {
			position85, tokenIndex85 := position, tokenIndex
			{
				position86 := position
				if !_rules[ruleident]() {
					goto l85
				}
				add(rulevariable_ref, position86)
			}
			return true
		l85:
			position, tokenIndex = position85, tokenIndex85
			return false
		},
		/* 20 ident <- <(([A-Z] / [a-z] / '_') ([A-Z] / [a-z] / [0-9] / '_')*)> */
		func() bool {
			position87, tokenIndex87 := position, tokenIndex
			{
				position88 := position
				{
					position89, tokenIndex89 := position, tokenIndex
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l90
					}
					position++
					goto l89
				l90:
					position, tokenIndex = position89, tokenIndex89
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l91
					}
					position++
					goto l89
				l91:
					position, tokenIndex = position89, tokenIndex89
					if buffer[position] != rune('_') {
						goto l87
					}
					position++
				}
			l89:
			l92:
				{
					position93, tokenIndex93 := position, tokenIndex
					{
						position94, tokenIndex94 := position, tokenIndex
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l95
						}
						position++
						goto l94
					l95:
						position, tokenIndex = position94, tokenIndex94
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l96
						}
						position++
						goto l94
					l96:
						position, tokenIndex = position94, tokenIndex94
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l97
						}
						position++
						goto l94
					l97:
						position, tokenIndex = position94, tokenIndex94
						if buffer[position] != rune('_') {
							goto l93
						}
						position++
					}
				l94:
					goto l92
				l93:
					position, tokenIndex = position93, tokenIndex93
				}
				add(ruleident, position88)
			}
			return true
		l87:
			position, tokenIndex = position87, tokenIndex87
			return false
		},
		/* 21 literal <- <(bool / float / int / string)> */
		func() bool {
			position98, tokenIndex98 := position, tokenIndex
			{
				position99 := position
				{
					position100, tokenIndex100 := position, tokenIndex
					if !_rules[rulebool]() {
						goto l101
					}
					goto l100
				l101:
					position, tokenIndex = position100, tokenIndex100
					if !_rules[rulefloat]() {
						goto l102
					}
					goto l100
				l102:
					position, tokenIndex = position100, tokenIndex100
					if !_rules[ruleint]() {
						goto l103
					}
					goto l100
				l103:
					position, tokenIndex = position100, tokenIndex100
					if !_rules[rulestring]() {
						goto l98
					}
				}
			l100:
				add(ruleliteral, position99)
			}
			return true
		l98:
			position, tokenIndex = position98, tokenIndex98
			return false
		},
		/* 22 float <- <(<('-'? [0-9]+ '.' [0-9]*)> Action19)> */
		func() bool {
			position104, tokenIndex104 := position, tokenIndex
			{
				position105 := position
				{
					position106 := position
					{
						position107, tokenIndex107 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l107
						}
						position++
						goto l108
					l107:
						position, tokenIndex = position107, tokenIndex107
					}
				l108:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l104
					}
					position++
				l109:
					{
						position110, tokenIndex110 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l110
						}
						position++
						goto l109
					l110:
						position, tokenIndex = position110, tokenIndex110
					}
					if buffer[position] != rune('.') {
						goto l104
					}
					position++
				l111:
					{
						position112, tokenIndex112 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l112
						}
						position++
						goto l111
					l112:
						position, tokenIndex = position112, tokenIndex112
					}
					add(rulePegText, position106)
				}
				if !_rules[ruleAction19]() {
					goto l104
				}
				add(rulefloat, position105)
			}
			return true
		l104:
			position, tokenIndex = position104, tokenIndex104
			return false
		},
		/* 23 int <- <(<('-'? [0-9]+)> Action20)> */
		func() bool {
			position113, tokenIndex113 := position, tokenIndex
			{
				position114 := position
				{
					position115 := position
					{
						position116, tokenIndex116 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l116
						}
						position++
						goto l117
					l116:
						position, tokenIndex = position116, tokenIndex116
					}
				l117:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l113
					}
					position++
				l118:
					{
						position119, tokenIndex119 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l119
						}
						position++
						goto l118
					l119:
						position, tokenIndex = position119, tokenIndex119
					}
					add(rulePegText, position115)
				}
				if !_rules[ruleAction20]() {
					goto l113
				}
				add(ruleint, position114)
			}
			return true
		l113:
			position, tokenIndex = position113, tokenIndex113
			return false
		},
		/* 24 bool <- <(<(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> Action21)> */
		func() bool {
			position120, tokenIndex120 := position, tokenIndex
			{
				position121 := position
				{
					position122 := position
					{
						position123, tokenIndex123 := position, tokenIndex
						if buffer[position] != rune('t') {
							goto l124
						}
						position++
						if buffer[position] != rune('r') {
							goto l124
						}
						position++
						if buffer[position] != rune('u') {
							goto l124
						}
						position++
						if buffer[position] != rune('e') {
							goto l124
						}
						position++
						goto l123
					l124:
						position, tokenIndex = position123, tokenIndex123
						if buffer[position] != rune('f') {
							goto l120
						}
						position++
						if buffer[position] != rune('a') {
							goto l120
						}
						position++
						if buffer[position] != rune('l') {
							goto l120
						}
						position++
						if buffer[position] != rune('s') {
							goto l120
						}
						position++
						if buffer[position] != rune('e') {
							goto l120
						}
						position++
					}
				l123:
					add(rulePegText, position122)
				}
				if !_rules[ruleAction21]() {
					goto l120
				}
				add(rulebool, position121)
			}
			return true
		l120:
			position, tokenIndex = position120, tokenIndex120
			return false
		},
		/* 25 string <- <(<('"' (!'"' .)* '"')> Action22)> */
		func() bool {
			position125, tokenIndex125 := position, tokenIndex
			{
				position126 := position
				{
					position127 := position
					if buffer[position] != rune('"') {
						goto l125
					}
					position++
				l128:
					{
						position129, tokenIndex129 := position, tokenIndex
						{
							position130, tokenIndex130 := position, tokenIndex
							if buffer[position] != rune('"') {
								goto l130
							}
							position++
							goto l129
						l130:
							position, tokenIndex = position130, tokenIndex130
						}
						if !matchDot() {
							goto l129
						}
						goto l128
					l129:
						position, tokenIndex = position129, tokenIndex129
					}
					if buffer[position] != rune('"') {
						goto l125
					}
					position++
					add(rulePegText, position127)
				}
				if !_rules[ruleAction22]() {
					goto l125
				}
				add(rulestring, position126)
			}
			return true
		l125:
			position, tokenIndex = position125, tokenIndex125
			return false
		},
		/* 26 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position132 := position
			l133:
				{
					position134, tokenIndex134 := position, tokenIndex
					{
						position135, tokenIndex135 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l136
						}
						position++
						goto l135
					l136:
						position, tokenIndex = position135, tokenIndex135
						if buffer[position] != rune('\t') {
							goto l134
						}
						position++
					}
				l135:
					goto l133
				l134:
					position, tokenIndex = position134, tokenIndex134
				}
				add(rulews, position132)
			}
			return true
		},
		/* 27 wsn <- <(' ' / '\t' / '\n')*> */
		func() bool {
			{
				position138 := position
			l139:
				{
					position140, tokenIndex140 := position, tokenIndex
					{
						position141, tokenIndex141 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l142
						}
						position++
						goto l141
					l142:
						position, tokenIndex = position141, tokenIndex141
						if buffer[position] != rune('\t') {
							goto l143
						}
						position++
						goto l141
					l143:
						position, tokenIndex = position141, tokenIndex141
						if buffer[position] != rune('\n') {
							goto l140
						}
						position++
					}
				l141:
					goto l139
				l140:
					position, tokenIndex = position140, tokenIndex140
				}
				add(rulewsn, position138)
			}
			return true
		},
		/* 29 Action0 <- <{
		    p.SetJump()
		}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 30 Action1 <- <{
		    p.AddCondJump()
		}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 31 Action2 <- <{
		    p.AddCode(vm.OpEMIT)
		}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 33 Action3 <- <{
		    p.AllocateRef(buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 34 Action4 <- <{
		    p.AddStore()
		}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 35 Action5 <- <{ p.AddCode(vm.OpAND) }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 36 Action6 <- <{ p.AddCode(vm.OpOR) }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 37 Action7 <- <{ p.AddCode(vm.OpEQ) }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 38 Action8 <- <{ p.AddCode(vm.OpNEQ) }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 39 Action9 <- <{ p.AddCode(vm.OpLT) }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 40 Action10 <- <{ p.AddCode(vm.OpLTE) }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 41 Action11 <- <{ p.AddCode(vm.OpGT) }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 42 Action12 <- <{ p.AddCode(vm.OpGTE) }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 43 Action13 <- <{ p.AddCode(vm.OpADD) }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 44 Action14 <- <{ p.AddCode(vm.OpSUB) }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 45 Action15 <- <{ p.AddCode(vm.OpMUL) }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 46 Action16 <- <{ p.AddCode(vm.OpDIV) }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 47 Action17 <- <{
		    p.AddCode(vm.OpCALL)
		}> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 48 Action18 <- <{
		    p.AddLoad(buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 49 Action19 <- <{
		    p.AddConstant(ParseFloat(buffer[begin:end]))
		}> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 50 Action20 <- <{
		    p.AddConstant(ParseInt(buffer[begin:end]))
		}> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 51 Action21 <- <{
		    p.AddConstant(ParseBool(buffer[begin:end]))
		}> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 52 Action22 <- <{
		    p.AddConstant(&vm.Str{buffer[begin:end]})
		}> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
	}
	p.rules = _rules
}
