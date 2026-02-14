package token

const (
	Undetermined Token = iota

	Skip

	Illegal
	Eof
	Comment

	String
	Number

	Plus      // +
	Minus     // -
	Multiply  // *
	Exponent  // **
	Slash     // /
	Remainder // %

	And                // &
	Or                 // |
	ExclusiveOr        // ^
	ShiftLeft          // <<
	ShiftRight         // >>
	UnsignedShiftRight // >>>

	AddAssign       // +=
	SubtractAssign  // -=
	MultiplyAssign  // *=
	ExponentAssign  // **=
	QuotientAssign  // /=
	RemainderAssign // %=

	AndAssign                // &=
	OrAssign                 // |=
	ExclusiveOrAssign        // ^=
	ShiftLeftAssign          // <<=
	ShiftRightAssign         // >>=
	UnsignedShiftRightAssign // >>>=

	LogicalAnd       // &&
	LogicalOr        // ||
	Coalesce         // ??
	LogicalAndAssign // &&=
	LogicalOrAssign  // ||=
	CoalesceAssign   // ??=
	Increment        // ++
	Decrement        // --

	Equal       // ==
	StrictEqual // ===
	Less        // <
	Greater     // >
	Assign      // =
	Not         // !

	BitwiseNot // ~

	NotEqual       // !=
	StrictNotEqual // !==
	LessOrEqual    // <=
	GreaterOrEqual // >=

	LeftParenthesis // (
	LeftBracket     // [
	LeftBrace       // {
	Comma           // ,
	Period          // .

	RightParenthesis // )
	RightBracket     // ]
	RightBrace       // }
	Semicolon        // ;
	Colon            // :
	QuestionMark     // ?
	QuestionDot      // ?.
	Arrow            // =>
	Ellipsis         // ...
	Backtick         // `

	PrivateIdentifier

	Identifier
	Keyword
	Boolean
	Null

	If
	In
	Of
	Do

	Var
	For
	New
	Try

	This
	Else
	Case
	Void
	With

	Const
	While
	Break
	Catch
	Throw
	Class
	Super

	Return
	Typeof
	Delete
	Switch

	Default
	Finally
	Extends

	Function
	Continue
	Debugger

	InstanceOf

	EscapedReservedWord

	Let
	Static
	Async
	Await
	Yield

	TemplateHead
	TemplateMiddle
	TemplateTail
	NoSubstitutionTemplate
)

var token2string = [...]string{
	Illegal:                  "Illegal",
	Eof:                      "Eof",
	Comment:                  "Comment",
	Keyword:                  "Keyword",
	String:                   "String",
	Boolean:                  "Boolean",
	Null:                     "Null",
	Number:                   "Number",
	Identifier:               "Identifier",
	PrivateIdentifier:        "PrivateIdentifier",
	Plus:                     "+",
	Minus:                    "-",
	Exponent:                 "**",
	Multiply:                 "*",
	Slash:                    "/",
	Remainder:                "%",
	And:                      "&",
	Or:                       "|",
	ExclusiveOr:              "^",
	ShiftLeft:                "<<",
	ShiftRight:               ">>",
	UnsignedShiftRight:       ">>>",
	AddAssign:                "+=",
	SubtractAssign:           "-=",
	MultiplyAssign:           "*=",
	ExponentAssign:           "**=",
	QuotientAssign:           "/=",
	RemainderAssign:          "%=",
	AndAssign:                "&=",
	OrAssign:                 "|=",
	ExclusiveOrAssign:        "^=",
	ShiftLeftAssign:          "<<=",
	ShiftRightAssign:         ">>=",
	UnsignedShiftRightAssign: ">>>=",
	LogicalAnd:               "&&",
	LogicalOr:                "||",
	Coalesce:                 "??",
	LogicalAndAssign:         "&&=",
	LogicalOrAssign:          "||=",
	CoalesceAssign:           "??=",
	Increment:                "++",
	Decrement:                "--",
	Equal:                    "==",
	StrictEqual:              "===",
	Less:                     "<",
	Greater:                  ">",
	Assign:                   "=",
	Not:                      "!",
	BitwiseNot:               "~",
	NotEqual:                 "!=",
	StrictNotEqual:           "!==",
	LessOrEqual:              "<=",
	GreaterOrEqual:           ">=",
	LeftParenthesis:          "(",
	LeftBracket:              "[",
	LeftBrace:                "{",
	Comma:                    ",",
	Period:                   ".",
	RightParenthesis:         ")",
	RightBracket:             "]",
	RightBrace:               "}",
	Semicolon:                ";",
	Colon:                    ":",
	QuestionMark:             "?",
	QuestionDot:              "?.",
	Arrow:                    "=>",
	Ellipsis:                 "...",
	Backtick:                 "`",
	If:                       "if",
	In:                       "in",
	Of:                       "of",
	Do:                       "do",
	Var:                      "var",
	Let:                      "let",
	For:                      "for",
	New:                      "new",
	Try:                      "try",
	This:                     "this",
	Else:                     "else",
	Case:                     "case",
	Void:                     "void",
	With:                     "with",
	Async:                    "async",
	Await:                    "await",
	Yield:                    "yield",
	Const:                    "const",
	While:                    "while",
	Break:                    "break",
	Catch:                    "catch",
	Throw:                    "throw",
	Class:                    "class",
	Super:                    "super",
	Return:                   "return",
	Typeof:                   "typeof",
	Delete:                   "delete",
	Switch:                   "switch",
	Static:                   "static",
	Default:                  "default",
	Finally:                  "finally",
	Extends:                  "extends",
	Function:                 "function",
	Continue:                 "continue",
	Debugger:                 "debugger",
	InstanceOf:               "instanceof",
}

var keywordTable = map[string]keyword{
	"if": {
		token: If,
	},
	"in": {
		token: In,
	},
	"do": {
		token: Do,
	},
	"var": {
		token: Var,
	},
	"for": {
		token: For,
	},
	"new": {
		token: New,
	},
	"try": {
		token: Try,
	},
	"this": {
		token: This,
	},
	"else": {
		token: Else,
	},
	"case": {
		token: Case,
	},
	"void": {
		token: Void,
	},
	"with": {
		token: With,
	},
	"async": {
		token: Async,
	},
	"while": {
		token: While,
	},
	"break": {
		token: Break,
	},
	"catch": {
		token: Catch,
	},
	"throw": {
		token: Throw,
	},
	"return": {
		token: Return,
	},
	"typeof": {
		token: Typeof,
	},
	"delete": {
		token: Delete,
	},
	"switch": {
		token: Switch,
	},
	"default": {
		token: Default,
	},
	"finally": {
		token: Finally,
	},
	"function": {
		token: Function,
	},
	"continue": {
		token: Continue,
	},
	"debugger": {
		token: Debugger,
	},
	"instanceof": {
		token: InstanceOf,
	},
	"const": {
		token: Const,
	},
	"class": {
		token: Class,
	},
	"enum": {
		token:         Keyword,
		futureKeyword: true,
	},
	"export": {
		token:         Keyword,
		futureKeyword: true,
	},
	"extends": {
		token: Extends,
	},
	"import": {
		token:         Keyword,
		futureKeyword: true,
	},
	"super": {
		token: Super,
	},
	"let": {
		token:  Let,
		strict: true,
	},
	"static": {
		token:  Static,
		strict: true,
	},
	"await": {
		token: Await,
	},
	"yield": {
		token: Yield,
	},
	"false": {
		token: Boolean,
	},
	"true": {
		token: Boolean,
	},
	"null": {
		token: Null,
	},
}
