package parse

const (
	TIdent = iota + 61601
	TNumber
	TString

	TFunc
	TReturn

	TIf
	TElse
	TElseIf

	TTrue
	TFalse

	TEqeq // ==
	TNeq  // !=
	TLte  // <=
	TGte  // >=

	TFor
	TBreak
	TContinue

	TAnd
	TOr
	TNot
	TIn

	TNil
)

// + - * / , { ( [ 等单字节的关键字 将ascii码值作为type值

var tokenName2Type = map[string]int{
	"func":     TFunc,
	"return":   TReturn,
	"if":       TIf,
	"else":     TElse,
	"elseif":   TElseIf,
	"true":     TTrue,
	"false":    TFalse,
	"for":      TFor,
	"break":    TBreak,
	"continue": TContinue,
	"and":      TAnd,
	"or":       TOr,
	"not":      TNot,
	"in":       TIn,
	"nil":      TNil,
}

var Type2TokenName = reverseMap(tokenName2Type)

func reverseMap(m map[string]int) map[int]string {
	n := make(map[int]string)
	for k, v := range m {
		n[v] = k
	}
	return n
}
