package vm

//go:generate peg vm/grammar.peg

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
	rulecomment
	ruleblock
	ruleif_block
	ruleif_cond
	ruleelse_block
	rulestmt
	ruleemit
	ruleassignment
	ruleglobal_assignment
	rulevar_assignment
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
	ruleAction23
	ruleAction24
)

var rul3s = [...]string{
	"Unknown",
	"file",
	"section",
	"comment",
	"block",
	"if_block",
	"if_cond",
	"else_block",
	"stmt",
	"emit",
	"assignment",
	"global_assignment",
	"var_assignment",
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
	"Action23",
	"Action24",
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
	constants     []Value
	code          []*Code
	jumpStack     []*Code
	registerCount int
	refs          map[string]int

	Buffer string
	buffer []rune
	rules  [59]func() bool
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

			p.AddCode(OpEMIT)

		case ruleAction3:

			p.PrepareMutate(buffer[begin:end])

		case ruleAction4:

			p.AddCode(OpMUT)

		case ruleAction5:

			p.AllocateRef(buffer[begin:end])

		case ruleAction6:

			p.AddStore()

		case ruleAction7:
			p.AddCondJump()
		case ruleAction8:
			p.SetJump()
		case ruleAction9:
			p.AddCode(OpEQ)
		case ruleAction10:
			p.AddCode(OpNEQ)
		case ruleAction11:
			p.AddCode(OpLT)
		case ruleAction12:
			p.AddCode(OpLTE)
		case ruleAction13:
			p.AddCode(OpGT)
		case ruleAction14:
			p.AddCode(OpGTE)
		case ruleAction15:
			p.AddCode(OpADD)
		case ruleAction16:
			p.AddCode(OpSUB)
		case ruleAction17:
			p.AddCode(OpMUL)
		case ruleAction18:
			p.AddCode(OpDIV)
		case ruleAction19:

			p.AddCode(OpCALL)

		case ruleAction20:

			p.AddLoad(buffer[begin:end])

		case ruleAction21:

			p.AddConstant(ParseFloat(buffer[begin:end]))

		case ruleAction22:

			p.AddConstant(ParseInt(buffer[begin:end]))

		case ruleAction23:

			p.AddConstant(ParseBool(buffer[begin:end]))

		case ruleAction24:

			p.AddConstant(&Str{buffer[begin:end]})

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
		/* 1 section <- <(comment / block / stmt)> */
		func() bool {
			position5, tokenIndex5 := position, tokenIndex
			{
				position6 := position
				{
					position7, tokenIndex7 := position, tokenIndex
					if !_rules[rulecomment]() {
						goto l8
					}
					goto l7
				l8:
					position, tokenIndex = position7, tokenIndex7
					if !_rules[ruleblock]() {
						goto l9
					}
					goto l7
				l9:
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
		/* 2 comment <- <('/' '/' (!'\n' .)* wsn)> */
		func() bool {
			position10, tokenIndex10 := position, tokenIndex
			{
				position11 := position
				if buffer[position] != rune('/') {
					goto l10
				}
				position++
				if buffer[position] != rune('/') {
					goto l10
				}
				position++
			l12:
				{
					position13, tokenIndex13 := position, tokenIndex
					{
						position14, tokenIndex14 := position, tokenIndex
						if buffer[position] != rune('\n') {
							goto l14
						}
						position++
						goto l13
					l14:
						position, tokenIndex = position14, tokenIndex14
					}
					if !matchDot() {
						goto l13
					}
					goto l12
				l13:
					position, tokenIndex = position13, tokenIndex13
				}
				if !_rules[rulewsn]() {
					goto l10
				}
				add(rulecomment, position11)
			}
			return true
		l10:
			position, tokenIndex = position10, tokenIndex10
			return false
		},
		/* 3 block <- <if_block> */
		func() bool {
			position15, tokenIndex15 := position, tokenIndex
			{
				position16 := position
				if !_rules[ruleif_block]() {
					goto l15
				}
				add(ruleblock, position16)
			}
			return true
		l15:
			position, tokenIndex = position15, tokenIndex15
			return false
		},
		/* 4 if_block <- <(if_cond '{' wsn section* wsn '}' wsn (else_block wsn)? Action0)> */
		func() bool {
			position17, tokenIndex17 := position, tokenIndex
			{
				position18 := position
				if !_rules[ruleif_cond]() {
					goto l17
				}
				if buffer[position] != rune('{') {
					goto l17
				}
				position++
				if !_rules[rulewsn]() {
					goto l17
				}
			l19:
				{
					position20, tokenIndex20 := position, tokenIndex
					if !_rules[rulesection]() {
						goto l20
					}
					goto l19
				l20:
					position, tokenIndex = position20, tokenIndex20
				}
				if !_rules[rulewsn]() {
					goto l17
				}
				if buffer[position] != rune('}') {
					goto l17
				}
				position++
				if !_rules[rulewsn]() {
					goto l17
				}
				{
					position21, tokenIndex21 := position, tokenIndex
					if !_rules[ruleelse_block]() {
						goto l21
					}
					if !_rules[rulewsn]() {
						goto l21
					}
					goto l22
				l21:
					position, tokenIndex = position21, tokenIndex21
				}
			l22:
				if !_rules[ruleAction0]() {
					goto l17
				}
				add(ruleif_block, position18)
			}
			return true
		l17:
			position, tokenIndex = position17, tokenIndex17
			return false
		},
		/* 5 if_cond <- <('i' 'f' ws '(' ws expr ws ')' wsn Action1)> */
		func() bool {
			position23, tokenIndex23 := position, tokenIndex
			{
				position24 := position
				if buffer[position] != rune('i') {
					goto l23
				}
				position++
				if buffer[position] != rune('f') {
					goto l23
				}
				position++
				if !_rules[rulews]() {
					goto l23
				}
				if buffer[position] != rune('(') {
					goto l23
				}
				position++
				if !_rules[rulews]() {
					goto l23
				}
				if !_rules[ruleexpr]() {
					goto l23
				}
				if !_rules[rulews]() {
					goto l23
				}
				if buffer[position] != rune(')') {
					goto l23
				}
				position++
				if !_rules[rulewsn]() {
					goto l23
				}
				if !_rules[ruleAction1]() {
					goto l23
				}
				add(ruleif_cond, position24)
			}
			return true
		l23:
			position, tokenIndex = position23, tokenIndex23
			return false
		},
		/* 6 else_block <- <('e' 'l' 's' 'e' wsn '{' wsn section* wsn '}' wsn)> */
		func() bool {
			position25, tokenIndex25 := position, tokenIndex
			{
				position26 := position
				if buffer[position] != rune('e') {
					goto l25
				}
				position++
				if buffer[position] != rune('l') {
					goto l25
				}
				position++
				if buffer[position] != rune('s') {
					goto l25
				}
				position++
				if buffer[position] != rune('e') {
					goto l25
				}
				position++
				if !_rules[rulewsn]() {
					goto l25
				}
				if buffer[position] != rune('{') {
					goto l25
				}
				position++
				if !_rules[rulewsn]() {
					goto l25
				}
			l27:
				{
					position28, tokenIndex28 := position, tokenIndex
					if !_rules[rulesection]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex = position28, tokenIndex28
				}
				if !_rules[rulewsn]() {
					goto l25
				}
				if buffer[position] != rune('}') {
					goto l25
				}
				position++
				if !_rules[rulewsn]() {
					goto l25
				}
				add(ruleelse_block, position26)
			}
			return true
		l25:
			position, tokenIndex = position25, tokenIndex25
			return false
		},
		/* 7 stmt <- <((emit / assignment) wsn)> */
		func() bool {
			position29, tokenIndex29 := position, tokenIndex
			{
				position30 := position
				{
					position31, tokenIndex31 := position, tokenIndex
					if !_rules[ruleemit]() {
						goto l32
					}
					goto l31
				l32:
					position, tokenIndex = position31, tokenIndex31
					if !_rules[ruleassignment]() {
						goto l29
					}
				}
			l31:
				if !_rules[rulewsn]() {
					goto l29
				}
				add(rulestmt, position30)
			}
			return true
		l29:
			position, tokenIndex = position29, tokenIndex29
			return false
		},
		/* 8 emit <- <('e' 'm' 'i' 't' ws expr Action2)> */
		func() bool {
			position33, tokenIndex33 := position, tokenIndex
			{
				position34 := position
				if buffer[position] != rune('e') {
					goto l33
				}
				position++
				if buffer[position] != rune('m') {
					goto l33
				}
				position++
				if buffer[position] != rune('i') {
					goto l33
				}
				position++
				if buffer[position] != rune('t') {
					goto l33
				}
				position++
				if !_rules[rulews]() {
					goto l33
				}
				if !_rules[ruleexpr]() {
					goto l33
				}
				if !_rules[ruleAction2]() {
					goto l33
				}
				add(ruleemit, position34)
			}
			return true
		l33:
			position, tokenIndex = position33, tokenIndex33
			return false
		},
		/* 9 assignment <- <(global_assignment / var_assignment)> */
		func() bool {
			position35, tokenIndex35 := position, tokenIndex
			{
				position36 := position
				{
					position37, tokenIndex37 := position, tokenIndex
					if !_rules[ruleglobal_assignment]() {
						goto l38
					}
					goto l37
				l38:
					position, tokenIndex = position37, tokenIndex37
					if !_rules[rulevar_assignment]() {
						goto l35
					}
				}
			l37:
				add(ruleassignment, position36)
			}
			return true
		l35:
			position, tokenIndex = position35, tokenIndex35
			return false
		},
		/* 10 global_assignment <- <(<global_ref> Action3 ws '=' ws expr Action4)> */
		func() bool {
			position39, tokenIndex39 := position, tokenIndex
			{
				position40 := position
				{
					position41 := position
					if !_rules[ruleglobal_ref]() {
						goto l39
					}
					add(rulePegText, position41)
				}
				if !_rules[ruleAction3]() {
					goto l39
				}
				if !_rules[rulews]() {
					goto l39
				}
				if buffer[position] != rune('=') {
					goto l39
				}
				position++
				if !_rules[rulews]() {
					goto l39
				}
				if !_rules[ruleexpr]() {
					goto l39
				}
				if !_rules[ruleAction4]() {
					goto l39
				}
				add(ruleglobal_assignment, position40)
			}
			return true
		l39:
			position, tokenIndex = position39, tokenIndex39
			return false
		},
		/* 11 var_assignment <- <(<variable_ref> Action5 ws '=' ws expr Action6)> */
		func() bool {
			position42, tokenIndex42 := position, tokenIndex
			{
				position43 := position
				{
					position44 := position
					if !_rules[rulevariable_ref]() {
						goto l42
					}
					add(rulePegText, position44)
				}
				if !_rules[ruleAction5]() {
					goto l42
				}
				if !_rules[rulews]() {
					goto l42
				}
				if buffer[position] != rune('=') {
					goto l42
				}
				position++
				if !_rules[rulews]() {
					goto l42
				}
				if !_rules[ruleexpr]() {
					goto l42
				}
				if !_rules[ruleAction6]() {
					goto l42
				}
				add(rulevar_assignment, position43)
			}
			return true
		l42:
			position, tokenIndex = position42, tokenIndex42
			return false
		},
		/* 12 expr <- <(or_expr / primary)> */
		func() bool {
			position45, tokenIndex45 := position, tokenIndex
			{
				position46 := position
				{
					position47, tokenIndex47 := position, tokenIndex
					if !_rules[ruleor_expr]() {
						goto l48
					}
					goto l47
				l48:
					position, tokenIndex = position47, tokenIndex47
					if !_rules[ruleprimary]() {
						goto l45
					}
				}
			l47:
				add(ruleexpr, position46)
			}
			return true
		l45:
			position, tokenIndex = position45, tokenIndex45
			return false
		},
		/* 13 primary <- <(('(' ws expr ws ')') / func_call / literal / deref)> */
		func() bool {
			position49, tokenIndex49 := position, tokenIndex
			{
				position50 := position
				{
					position51, tokenIndex51 := position, tokenIndex
					if buffer[position] != rune('(') {
						goto l52
					}
					position++
					if !_rules[rulews]() {
						goto l52
					}
					if !_rules[ruleexpr]() {
						goto l52
					}
					if !_rules[rulews]() {
						goto l52
					}
					if buffer[position] != rune(')') {
						goto l52
					}
					position++
					goto l51
				l52:
					position, tokenIndex = position51, tokenIndex51
					if !_rules[rulefunc_call]() {
						goto l53
					}
					goto l51
				l53:
					position, tokenIndex = position51, tokenIndex51
					if !_rules[ruleliteral]() {
						goto l54
					}
					goto l51
				l54:
					position, tokenIndex = position51, tokenIndex51
					if !_rules[rulederef]() {
						goto l49
					}
				}
			l51:
				add(ruleprimary, position50)
			}
			return true
		l49:
			position, tokenIndex = position49, tokenIndex49
			return false
		},
		/* 14 or_expr <- <(and_expr (ws ('|' '|') ws expr)*)> */
		func() bool {
			position55, tokenIndex55 := position, tokenIndex
			{
				position56 := position
				if !_rules[ruleand_expr]() {
					goto l55
				}
			l57:
				{
					position58, tokenIndex58 := position, tokenIndex
					if !_rules[rulews]() {
						goto l58
					}
					if buffer[position] != rune('|') {
						goto l58
					}
					position++
					if buffer[position] != rune('|') {
						goto l58
					}
					position++
					if !_rules[rulews]() {
						goto l58
					}
					if !_rules[ruleexpr]() {
						goto l58
					}
					goto l57
				l58:
					position, tokenIndex = position58, tokenIndex58
				}
				add(ruleor_expr, position56)
			}
			return true
		l55:
			position, tokenIndex = position55, tokenIndex55
			return false
		},
		/* 15 and_expr <- <(compare_expr (ws ('&' '&') Action7 ws expr Action8)*)> */
		func() bool {
			position59, tokenIndex59 := position, tokenIndex
			{
				position60 := position
				if !_rules[rulecompare_expr]() {
					goto l59
				}
			l61:
				{
					position62, tokenIndex62 := position, tokenIndex
					if !_rules[rulews]() {
						goto l62
					}
					if buffer[position] != rune('&') {
						goto l62
					}
					position++
					if buffer[position] != rune('&') {
						goto l62
					}
					position++
					if !_rules[ruleAction7]() {
						goto l62
					}
					if !_rules[rulews]() {
						goto l62
					}
					if !_rules[ruleexpr]() {
						goto l62
					}
					if !_rules[ruleAction8]() {
						goto l62
					}
					goto l61
				l62:
					position, tokenIndex = position62, tokenIndex62
				}
				add(ruleand_expr, position60)
			}
			return true
		l59:
			position, tokenIndex = position59, tokenIndex59
			return false
		},
		/* 16 compare_expr <- <(add_expr (ws (('=' '=' ws expr Action9) / ('!' '=' ws expr Action10) / ('<' ws expr Action11) / ('<' '=' ws expr Action12) / ('>' ws expr Action13) / ('>' '=' ws expr Action14)))*)> */
		func() bool {
			position63, tokenIndex63 := position, tokenIndex
			{
				position64 := position
				if !_rules[ruleadd_expr]() {
					goto l63
				}
			l65:
				{
					position66, tokenIndex66 := position, tokenIndex
					if !_rules[rulews]() {
						goto l66
					}
					{
						position67, tokenIndex67 := position, tokenIndex
						if buffer[position] != rune('=') {
							goto l68
						}
						position++
						if buffer[position] != rune('=') {
							goto l68
						}
						position++
						if !_rules[rulews]() {
							goto l68
						}
						if !_rules[ruleexpr]() {
							goto l68
						}
						if !_rules[ruleAction9]() {
							goto l68
						}
						goto l67
					l68:
						position, tokenIndex = position67, tokenIndex67
						if buffer[position] != rune('!') {
							goto l69
						}
						position++
						if buffer[position] != rune('=') {
							goto l69
						}
						position++
						if !_rules[rulews]() {
							goto l69
						}
						if !_rules[ruleexpr]() {
							goto l69
						}
						if !_rules[ruleAction10]() {
							goto l69
						}
						goto l67
					l69:
						position, tokenIndex = position67, tokenIndex67
						if buffer[position] != rune('<') {
							goto l70
						}
						position++
						if !_rules[rulews]() {
							goto l70
						}
						if !_rules[ruleexpr]() {
							goto l70
						}
						if !_rules[ruleAction11]() {
							goto l70
						}
						goto l67
					l70:
						position, tokenIndex = position67, tokenIndex67
						if buffer[position] != rune('<') {
							goto l71
						}
						position++
						if buffer[position] != rune('=') {
							goto l71
						}
						position++
						if !_rules[rulews]() {
							goto l71
						}
						if !_rules[ruleexpr]() {
							goto l71
						}
						if !_rules[ruleAction12]() {
							goto l71
						}
						goto l67
					l71:
						position, tokenIndex = position67, tokenIndex67
						if buffer[position] != rune('>') {
							goto l72
						}
						position++
						if !_rules[rulews]() {
							goto l72
						}
						if !_rules[ruleexpr]() {
							goto l72
						}
						if !_rules[ruleAction13]() {
							goto l72
						}
						goto l67
					l72:
						position, tokenIndex = position67, tokenIndex67
						if buffer[position] != rune('>') {
							goto l66
						}
						position++
						if buffer[position] != rune('=') {
							goto l66
						}
						position++
						if !_rules[rulews]() {
							goto l66
						}
						if !_rules[ruleexpr]() {
							goto l66
						}
						if !_rules[ruleAction14]() {
							goto l66
						}
					}
				l67:
					goto l65
				l66:
					position, tokenIndex = position66, tokenIndex66
				}
				add(rulecompare_expr, position64)
			}
			return true
		l63:
			position, tokenIndex = position63, tokenIndex63
			return false
		},
		/* 17 add_expr <- <(mult_expr (ws (('+' ws expr Action15) / ('-' ws expr Action16)))*)> */
		func() bool {
			position73, tokenIndex73 := position, tokenIndex
			{
				position74 := position
				if !_rules[rulemult_expr]() {
					goto l73
				}
			l75:
				{
					position76, tokenIndex76 := position, tokenIndex
					if !_rules[rulews]() {
						goto l76
					}
					{
						position77, tokenIndex77 := position, tokenIndex
						if buffer[position] != rune('+') {
							goto l78
						}
						position++
						if !_rules[rulews]() {
							goto l78
						}
						if !_rules[ruleexpr]() {
							goto l78
						}
						if !_rules[ruleAction15]() {
							goto l78
						}
						goto l77
					l78:
						position, tokenIndex = position77, tokenIndex77
						if buffer[position] != rune('-') {
							goto l76
						}
						position++
						if !_rules[rulews]() {
							goto l76
						}
						if !_rules[ruleexpr]() {
							goto l76
						}
						if !_rules[ruleAction16]() {
							goto l76
						}
					}
				l77:
					goto l75
				l76:
					position, tokenIndex = position76, tokenIndex76
				}
				add(ruleadd_expr, position74)
			}
			return true
		l73:
			position, tokenIndex = position73, tokenIndex73
			return false
		},
		/* 18 mult_expr <- <(primary (ws (('*' ws expr Action17) / ('/' ws expr Action18)))*)> */
		func() bool {
			position79, tokenIndex79 := position, tokenIndex
			{
				position80 := position
				if !_rules[ruleprimary]() {
					goto l79
				}
			l81:
				{
					position82, tokenIndex82 := position, tokenIndex
					if !_rules[rulews]() {
						goto l82
					}
					{
						position83, tokenIndex83 := position, tokenIndex
						if buffer[position] != rune('*') {
							goto l84
						}
						position++
						if !_rules[rulews]() {
							goto l84
						}
						if !_rules[ruleexpr]() {
							goto l84
						}
						if !_rules[ruleAction17]() {
							goto l84
						}
						goto l83
					l84:
						position, tokenIndex = position83, tokenIndex83
						if buffer[position] != rune('/') {
							goto l82
						}
						position++
						if !_rules[rulews]() {
							goto l82
						}
						if !_rules[ruleexpr]() {
							goto l82
						}
						if !_rules[ruleAction18]() {
							goto l82
						}
					}
				l83:
					goto l81
				l82:
					position, tokenIndex = position82, tokenIndex82
				}
				add(rulemult_expr, position80)
			}
			return true
		l79:
			position, tokenIndex = position79, tokenIndex79
			return false
		},
		/* 19 func_call <- <(deref '(' wsn (expr (wsn ',' wsn expr)*)? wsn ')' Action19)> */
		func() bool {
			position85, tokenIndex85 := position, tokenIndex
			{
				position86 := position
				if !_rules[rulederef]() {
					goto l85
				}
				if buffer[position] != rune('(') {
					goto l85
				}
				position++
				if !_rules[rulewsn]() {
					goto l85
				}
				{
					position87, tokenIndex87 := position, tokenIndex
					if !_rules[ruleexpr]() {
						goto l87
					}
				l89:
					{
						position90, tokenIndex90 := position, tokenIndex
						if !_rules[rulewsn]() {
							goto l90
						}
						if buffer[position] != rune(',') {
							goto l90
						}
						position++
						if !_rules[rulewsn]() {
							goto l90
						}
						if !_rules[ruleexpr]() {
							goto l90
						}
						goto l89
					l90:
						position, tokenIndex = position90, tokenIndex90
					}
					goto l88
				l87:
					position, tokenIndex = position87, tokenIndex87
				}
			l88:
				if !_rules[rulewsn]() {
					goto l85
				}
				if buffer[position] != rune(')') {
					goto l85
				}
				position++
				if !_rules[ruleAction19]() {
					goto l85
				}
				add(rulefunc_call, position86)
			}
			return true
		l85:
			position, tokenIndex = position85, tokenIndex85
			return false
		},
		/* 20 ref <- <(global_ref / variable_ref)> */
		func() bool {
			position91, tokenIndex91 := position, tokenIndex
			{
				position92 := position
				{
					position93, tokenIndex93 := position, tokenIndex
					if !_rules[ruleglobal_ref]() {
						goto l94
					}
					goto l93
				l94:
					position, tokenIndex = position93, tokenIndex93
					if !_rules[rulevariable_ref]() {
						goto l91
					}
				}
			l93:
				add(ruleref, position92)
			}
			return true
		l91:
			position, tokenIndex = position91, tokenIndex91
			return false
		},
		/* 21 deref <- <(<ref> Action20)> */
		func() bool {
			position95, tokenIndex95 := position, tokenIndex
			{
				position96 := position
				{
					position97 := position
					if !_rules[ruleref]() {
						goto l95
					}
					add(rulePegText, position97)
				}
				if !_rules[ruleAction20]() {
					goto l95
				}
				add(rulederef, position96)
			}
			return true
		l95:
			position, tokenIndex = position95, tokenIndex95
			return false
		},
		/* 22 global_ref <- <('@' (ident ('.' ident)*)?)> */
		func() bool {
			position98, tokenIndex98 := position, tokenIndex
			{
				position99 := position
				if buffer[position] != rune('@') {
					goto l98
				}
				position++
				{
					position100, tokenIndex100 := position, tokenIndex
					if !_rules[ruleident]() {
						goto l100
					}
				l102:
					{
						position103, tokenIndex103 := position, tokenIndex
						if buffer[position] != rune('.') {
							goto l103
						}
						position++
						if !_rules[ruleident]() {
							goto l103
						}
						goto l102
					l103:
						position, tokenIndex = position103, tokenIndex103
					}
					goto l101
				l100:
					position, tokenIndex = position100, tokenIndex100
				}
			l101:
				add(ruleglobal_ref, position99)
			}
			return true
		l98:
			position, tokenIndex = position98, tokenIndex98
			return false
		},
		/* 23 variable_ref <- <ident> */
		func() bool {
			position104, tokenIndex104 := position, tokenIndex
			{
				position105 := position
				if !_rules[ruleident]() {
					goto l104
				}
				add(rulevariable_ref, position105)
			}
			return true
		l104:
			position, tokenIndex = position104, tokenIndex104
			return false
		},
		/* 24 ident <- <(([A-Z] / [a-z] / '_') ([A-Z] / [a-z] / [0-9] / '_')*)> */
		func() bool {
			position106, tokenIndex106 := position, tokenIndex
			{
				position107 := position
				{
					position108, tokenIndex108 := position, tokenIndex
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l109
					}
					position++
					goto l108
				l109:
					position, tokenIndex = position108, tokenIndex108
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l110
					}
					position++
					goto l108
				l110:
					position, tokenIndex = position108, tokenIndex108
					if buffer[position] != rune('_') {
						goto l106
					}
					position++
				}
			l108:
			l111:
				{
					position112, tokenIndex112 := position, tokenIndex
					{
						position113, tokenIndex113 := position, tokenIndex
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l114
						}
						position++
						goto l113
					l114:
						position, tokenIndex = position113, tokenIndex113
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l115
						}
						position++
						goto l113
					l115:
						position, tokenIndex = position113, tokenIndex113
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l116
						}
						position++
						goto l113
					l116:
						position, tokenIndex = position113, tokenIndex113
						if buffer[position] != rune('_') {
							goto l112
						}
						position++
					}
				l113:
					goto l111
				l112:
					position, tokenIndex = position112, tokenIndex112
				}
				add(ruleident, position107)
			}
			return true
		l106:
			position, tokenIndex = position106, tokenIndex106
			return false
		},
		/* 25 literal <- <(bool / float / int / string)> */
		func() bool {
			position117, tokenIndex117 := position, tokenIndex
			{
				position118 := position
				{
					position119, tokenIndex119 := position, tokenIndex
					if !_rules[rulebool]() {
						goto l120
					}
					goto l119
				l120:
					position, tokenIndex = position119, tokenIndex119
					if !_rules[rulefloat]() {
						goto l121
					}
					goto l119
				l121:
					position, tokenIndex = position119, tokenIndex119
					if !_rules[ruleint]() {
						goto l122
					}
					goto l119
				l122:
					position, tokenIndex = position119, tokenIndex119
					if !_rules[rulestring]() {
						goto l117
					}
				}
			l119:
				add(ruleliteral, position118)
			}
			return true
		l117:
			position, tokenIndex = position117, tokenIndex117
			return false
		},
		/* 26 float <- <(<('-'? [0-9]+ '.' [0-9]*)> Action21)> */
		func() bool {
			position123, tokenIndex123 := position, tokenIndex
			{
				position124 := position
				{
					position125 := position
					{
						position126, tokenIndex126 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l126
						}
						position++
						goto l127
					l126:
						position, tokenIndex = position126, tokenIndex126
					}
				l127:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l123
					}
					position++
				l128:
					{
						position129, tokenIndex129 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l129
						}
						position++
						goto l128
					l129:
						position, tokenIndex = position129, tokenIndex129
					}
					if buffer[position] != rune('.') {
						goto l123
					}
					position++
				l130:
					{
						position131, tokenIndex131 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l131
						}
						position++
						goto l130
					l131:
						position, tokenIndex = position131, tokenIndex131
					}
					add(rulePegText, position125)
				}
				if !_rules[ruleAction21]() {
					goto l123
				}
				add(rulefloat, position124)
			}
			return true
		l123:
			position, tokenIndex = position123, tokenIndex123
			return false
		},
		/* 27 int <- <(<('-'? [0-9]+)> Action22)> */
		func() bool {
			position132, tokenIndex132 := position, tokenIndex
			{
				position133 := position
				{
					position134 := position
					{
						position135, tokenIndex135 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l135
						}
						position++
						goto l136
					l135:
						position, tokenIndex = position135, tokenIndex135
					}
				l136:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l132
					}
					position++
				l137:
					{
						position138, tokenIndex138 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l138
						}
						position++
						goto l137
					l138:
						position, tokenIndex = position138, tokenIndex138
					}
					add(rulePegText, position134)
				}
				if !_rules[ruleAction22]() {
					goto l132
				}
				add(ruleint, position133)
			}
			return true
		l132:
			position, tokenIndex = position132, tokenIndex132
			return false
		},
		/* 28 bool <- <(<(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> Action23)> */
		func() bool {
			position139, tokenIndex139 := position, tokenIndex
			{
				position140 := position
				{
					position141 := position
					{
						position142, tokenIndex142 := position, tokenIndex
						if buffer[position] != rune('t') {
							goto l143
						}
						position++
						if buffer[position] != rune('r') {
							goto l143
						}
						position++
						if buffer[position] != rune('u') {
							goto l143
						}
						position++
						if buffer[position] != rune('e') {
							goto l143
						}
						position++
						goto l142
					l143:
						position, tokenIndex = position142, tokenIndex142
						if buffer[position] != rune('f') {
							goto l139
						}
						position++
						if buffer[position] != rune('a') {
							goto l139
						}
						position++
						if buffer[position] != rune('l') {
							goto l139
						}
						position++
						if buffer[position] != rune('s') {
							goto l139
						}
						position++
						if buffer[position] != rune('e') {
							goto l139
						}
						position++
					}
				l142:
					add(rulePegText, position141)
				}
				if !_rules[ruleAction23]() {
					goto l139
				}
				add(rulebool, position140)
			}
			return true
		l139:
			position, tokenIndex = position139, tokenIndex139
			return false
		},
		/* 29 string <- <('"' <(!'"' .)*> '"' Action24)> */
		func() bool {
			position144, tokenIndex144 := position, tokenIndex
			{
				position145 := position
				if buffer[position] != rune('"') {
					goto l144
				}
				position++
				{
					position146 := position
				l147:
					{
						position148, tokenIndex148 := position, tokenIndex
						{
							position149, tokenIndex149 := position, tokenIndex
							if buffer[position] != rune('"') {
								goto l149
							}
							position++
							goto l148
						l149:
							position, tokenIndex = position149, tokenIndex149
						}
						if !matchDot() {
							goto l148
						}
						goto l147
					l148:
						position, tokenIndex = position148, tokenIndex148
					}
					add(rulePegText, position146)
				}
				if buffer[position] != rune('"') {
					goto l144
				}
				position++
				if !_rules[ruleAction24]() {
					goto l144
				}
				add(rulestring, position145)
			}
			return true
		l144:
			position, tokenIndex = position144, tokenIndex144
			return false
		},
		/* 30 ws <- <(' ' / '\t')*> */
		func() bool {
			{
				position151 := position
			l152:
				{
					position153, tokenIndex153 := position, tokenIndex
					{
						position154, tokenIndex154 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l155
						}
						position++
						goto l154
					l155:
						position, tokenIndex = position154, tokenIndex154
						if buffer[position] != rune('\t') {
							goto l153
						}
						position++
					}
				l154:
					goto l152
				l153:
					position, tokenIndex = position153, tokenIndex153
				}
				add(rulews, position151)
			}
			return true
		},
		/* 31 wsn <- <(' ' / '\t' / '\n')*> */
		func() bool {
			{
				position157 := position
			l158:
				{
					position159, tokenIndex159 := position, tokenIndex
					{
						position160, tokenIndex160 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l161
						}
						position++
						goto l160
					l161:
						position, tokenIndex = position160, tokenIndex160
						if buffer[position] != rune('\t') {
							goto l162
						}
						position++
						goto l160
					l162:
						position, tokenIndex = position160, tokenIndex160
						if buffer[position] != rune('\n') {
							goto l159
						}
						position++
					}
				l160:
					goto l158
				l159:
					position, tokenIndex = position159, tokenIndex159
				}
				add(rulewsn, position157)
			}
			return true
		},
		/* 33 Action0 <- <{
		    p.SetJump()
		}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 34 Action1 <- <{
		    p.AddCondJump()
		}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 35 Action2 <- <{
		    p.AddCode(OpEMIT)
		}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 37 Action3 <- <{
		    p.PrepareMutate(buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 38 Action4 <- <{
		    p.AddCode(OpMUT)
		}> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 39 Action5 <- <{
		    p.AllocateRef(buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 40 Action6 <- <{
		    p.AddStore()
		}> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 41 Action7 <- <{ p.AddCondJump() }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 42 Action8 <- <{ p.SetJump() }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 43 Action9 <- <{ p.AddCode(OpEQ) }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 44 Action10 <- <{ p.AddCode(OpNEQ) }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 45 Action11 <- <{ p.AddCode(OpLT) }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 46 Action12 <- <{ p.AddCode(OpLTE) }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 47 Action13 <- <{ p.AddCode(OpGT) }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 48 Action14 <- <{ p.AddCode(OpGTE) }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 49 Action15 <- <{ p.AddCode(OpADD) }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 50 Action16 <- <{ p.AddCode(OpSUB) }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 51 Action17 <- <{ p.AddCode(OpMUL) }> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 52 Action18 <- <{ p.AddCode(OpDIV) }> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 53 Action19 <- <{
		    p.AddCode(OpCALL)
		}> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 54 Action20 <- <{
		    p.AddLoad(buffer[begin:end])
		}> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 55 Action21 <- <{
		    p.AddConstant(ParseFloat(buffer[begin:end]))
		}> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 56 Action22 <- <{
		    p.AddConstant(ParseInt(buffer[begin:end]))
		}> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 57 Action23 <- <{
		    p.AddConstant(ParseBool(buffer[begin:end]))
		}> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 58 Action24 <- <{
		    p.AddConstant(&Str{buffer[begin:end]})
		}> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
	}
	p.rules = _rules
}
