package main

// util.go

// Vector
type Vector struct {
	data     []interface{}
	capacity int
	len      int
}

// Map
type Map struct {
	keys *Vector
	vals *Vector
}

// StringBuilder
type StringBuilder struct {
	data     string
	capacity int
	len      int
}

type Type struct {
	ty    int
	size  int // sizeof
	align int // alignof

	// Pointer
	ptr_to *Type

	// Array
	ary_of *Type
	len    int

	// Struct
	members *Vector
	offset  int

	// Function
	returning *Type
}

// token.go

const (
	TK_NUM       = iota + 256 // Number literal
	TK_STR                    // String literal
	TK_IDENT                  // Identifier
	TK_ARROW                  // ->
	TK_EXTERN                 // "extern"
	TK_TYPEDEF                // "typedef"
	TK_INT                    // "int"
	TK_CHAR                   // "char"
	TK_VOID                   // "void"
	TK_STRUCT                 // "struct"
	TK_IF                     // "if"
	TK_ELSE                   // "else"
	TK_FOR                    // "for"
	TK_DO                     // "do"
	TK_WHILE                  // "while"
	TK_BREAK                  // "break"
	TK_EQ                     // ==
	TK_NE                     // !=
	TK_LE                     // <=
	TK_GE                     // >=
	TK_LOGOR                  // ||
	TK_LOGAND                 // &&
	TK_SHL                    // <<
	TK_SHR                    // >>
	TK_INC                    // ++
	TK_DEC                    // --
	TK_MUL_EQ                 // *=
	TK_DIV_EQ                 // /=
	TK_MOD_EQ                 // %=
	TK_ADD_EQ                 // +=
	TK_SUB_EQ                 // -=
	TK_SHL_EQ                 // <<=
	TK_SHR_EQ                 // >>=
	TK_BITAND_EQ              // &=
	TK_XOR_EQ                 // ^=
	TK_BITOR_EQ               // |=
	TK_RETURN                 // "return"
	TK_SIZEOF                 // "sizeof"
	TK_ALIGNOF                // "_Alignof"
	TK_PARAM                  // Function-like macro parameter
	TK_EOF                    // End marker
)

// Token type
type Token struct {
	ty   int    // Token type
	val  int    // Number literal
	name string // Identifier

	// String literal
	str string
	len int

	// For preprocessor
	stringize bool

	// For error reporting
	buf   string
	path  string
	start string
	end   string
}

// parse.go
const (
	ND_NUM       = iota + 256 // Number literal
	ND_STR                    // String literal
	ND_IDENT                  // Identigier
	ND_STRUCT                 // Struct
	ND_DECL                   // declaration
	ND_VARDEF                 // Variable definition
	ND_LVAR                   // Local variable reference
	ND_GVAR                   // Global variable reference
	ND_IF                     // "if"
	ND_FOR                    // "for"
	ND_DO_WHILE               // do ... while
	ND_BREAK                  // break
	ND_ADDR                   // address-of operator ("&")
	ND_DEREF                  // pointer dereference ("*")
	ND_DOT                    // Struct member access
	ND_EQ                     // ==
	ND_NE                     // !=
	ND_LE                     // <=
	ND_LOGOR                  // ||
	ND_LOGAND                 // &&
	ND_SHL                    // <<
	ND_SHR                    // >>
	ND_MOD                    // %
	ND_NEG                    // -
	ND_POST_INC               // post ++
	ND_POST_DEC               // post --
	ND_MUL_EQ                 // *=
	ND_DIV_EQ                 // /=
	ND_MOD_EQ                 // %=
	ND_ADD_EQ                 // +=
	ND_SUB_EQ                 // -=
	ND_SHL_EQ                 // <<=
	ND_SHR_EQ                 // >>=
	ND_BITAND_EQ              // &=
	ND_XOR_EQ                 // ^=
	ND_BITOR_EQ               // |=
	ND_RETURN                 // "return"
	ND_SIZEOF                 // "sizeof"
	ND_ALIGNOF                // "_Alignof"
	ND_CALL                   // Function call
	ND_FUNC                   // Function definition
	ND_COMP_STMT              // Compound statement
	ND_EXPR_STMT              // Expressions statement
	ND_STMT_EXPR              // Statement expression (GUN extn.)
	ND_NULL                   // Null statement
)

const (
	INT = iota
	CHAR
	VOID
	PTR
	ARY
	STRUCT
	FUNC
)

type Node struct {
	op    int     // Node type
	ty    *Type   // C type
	lhs   *Node   // left-hand side
	rhs   *Node   // right-hand side
	val   int     // Number literal
	expr  *Node   // "return" or expression stmt
	stmts *Vector // Compound statement

	name string // Identifier

	// Global variable
	is_extern bool
	data      string
	len       int

	// "if" ( cond ) then "else" els
	// "for" ( init; cond; inc ) body
	cond *Node
	then *Node
	els  *Node
	init *Node
	body *Node
	inc  *Node

	// Function definition
	stacksize int
	globals   *Vector

	// Offset from BP or beginning of a struct
	offset int

	// Function call
	args *Vector
}

// sema.go

type Var struct {
	ty       *Type
	is_local bool

	// local
	offset int

	// global
	name      string
	is_extern bool
	data      string
	len       int
}

// ir_dump.go

type IRInfo struct {
	name string
	ty   int
}

// gen_ir.go

const (
	IR_ADD = iota + 256
	IR_SUB
	IR_MUL
	IR_DIV
	IR_IMM
	IR_BPREL
	IR_MOV
	IR_RETURN
	IR_CALL
	IR_LABEL
	IR_LABEL_ADDR
	IR_EQ
	IR_NE
	IR_LE
	IR_LT
	IR_AND
	IR_OR
	IR_XOR
	IR_SHL
	IR_SHR
	IR_MOD
	IR_NEG
	IR_JMP
	IR_IF
	IR_UNLESS
	IR_LOAD
	IR_STORE
	IR_STORE_ARG
	IR_KILL
	IR_NOP
)

type IR struct {
	op  int
	lhs int
	rhs int

	// Load/Store size in bytes
	size int

	// For binary operator. If true, rhs is an immediate.
	is_imm bool

	// Function call
	name  string
	nargs int
	args  [6]int
}

const (
	IR_TY_NOARG = iota + 256
	IR_TY_BINARY
	IR_TY_REG
	IR_TY_IMM
	IR_TY_MEM
	IR_TY_JMP
	IR_TY_LABEL
	IR_TY_LABEL_ADDR
	IR_TY_REG_REG
	IR_TY_REG_IMM
	IR_TY_STORE_ARG
	IR_TY_REG_LABEL
	IR_TY_CALL
)

type Function struct {
	name      string
	stacksize int
	globals   *Vector
	ir        *Vector
}
