package tagvalue

import (
	"fmt"
	"html/template"
	"strings"
)

// Sanity check DKIM keys for compliance with RFCs 6376,
// 8301, 8463, 8553 and 8616

// DkimKey represents a DKIM public key, as published to DNS
// in a format intended for diagnostics and display
type DkimKey struct {
	V Field
	G Field
	H Field
	K Field
	N Field
	P Field
	S Field
	T Field
	Unrecognized map[string]Field
	ParseError ParseError
}

func NewDkimKey(input string) DkimKey {
	Fields, err := NewMap(input)
	if err != nil {
		switch v := err.(type) {
		case ParseError:
			return DkimKey{ParseError: v}
		default:
			return DkimKey{ParseError:ParseError{Message:err.Error()}}
		}
	}

	ret := DkimKey{
		V: Fields["v"],
		G: Fields["g"],
		H: Fields["h"],
		K: Fields["k"],
		N: Fields["n"],
		P: Fields["p"],
		S: Fields["s"],
		T: Fields["t"],
		Unrecognized: Fields,
	}
	for _, k := range []string{"v", "g", "h", "k", "n", "p", "s", "t"} {
		delete(ret.Unrecognized, k)
	}

	// Annotate version
	if ret.V.Defined {
		if ret.V.Value != "DKIM1" {
			ret.V.addError(template.HTML(`The version field must be <a href="/rfc/6376#section-3.6.1">DKIM1</a>"`))
		}
		if ret.V.Index != 0 {
			ret.V.addError(template.HTML(`The version tag must be the <a href="https://tools.wordtothewise.com/rfc/6376#section-3.6.1">first tag in the record</a>`))
		}
	} else {
		ret.V.addWarning("DKIM key records should ideally have a version field")
	}

	// Annotate granularity
	if ret.G.Defined {
		if ret.G.Value == "*" {
			ret.V.addWarning(template.HTML(`The granularity field ("g=*") is deprecated in <a href="https://tools.wordtothewise.com/rfc/6376#appendix-C.2">RFC 6376</a>` ))
		} else {
			ret.V.addError(`The granularity field ("g=") is deprecated in <a href="https://tools.wordtothewise.com/rfc/6376#appendix-C.2">RFC 6376</a> and this value will be treated differently by pre-6376 and post-6376 validators`)
		}
	}

	// Annotate hash
	if ret.H.Defined {
		algos := strings.Split(ret.H.Value, ":")
		for _, algo := range algos {
			switch algo {
			case "sha256":
			case "sha1":
				ret.H.addWarning(template.HTML(`SHA1 is <a href="/rfc/8301#section-3.1">not a trusted hash</a>, mail using it may fail DKIM now or in the future`))
			default:
				ret.H.addWarning(fmt.Sprintf("'%s' isn't a hash type I recognize"))
			}
		}
	}

	// Annotate signing algorithm
	if ret.K.Defined {
		switch ret.K.Value {

		}
	}
}
