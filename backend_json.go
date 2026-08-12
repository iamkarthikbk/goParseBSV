// Copyright (c) 2016-2017 Rishiyur Nikhil and Bluespec, Inc.  All Rights Reserved.

// backend_json.go: emit the AST as JSON, for consumption by other tools.
//
// Reflective rather than a case per node type: AST is an empty interface and node
// fields are exported, so one walker covers every node type. Each node becomes
// {"kind": "AstRule", ...fields}; a *Token becomes
// {"kind": "Token", "tok": ..., "s": ..., "line": ..., "col": ...}.

package goParseBSV

import (
	"encoding/json"
	"os"
	"reflect"
)

func tokTypeName(t uint) string {
	switch t {
	case TokNone:
		return "None"
	case TokEof:
		return "Eof"
	case TokKeyword:
		return "Keyword"
	case TokIde:
		return "Ide"
	case TokInteger:
		return "Integer"
	case TokString:
		return "String"
	default:
		return "Other"
	}
}

// astToPlain converts an AST node into values encoding/json understands.
func astToPlain(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		if tok, ok := v.Interface().(*Token); ok {
			return map[string]interface{}{
				"kind":  "Token",
				"tok":   tokTypeName(tok.TokType),
				"s":     tok.StringVal,
				"i":     tok.IntVal,
				"width": tok.IntWidth,
				"line":  tok.LineNum,
				"col":   tok.Column,
			}
		}
		return astToPlain(v.Elem())

	case reflect.Slice, reflect.Array:
		out := make([]interface{}, 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			out = append(out, astToPlain(v.Index(i)))
		}
		return out

	case reflect.Struct:
		out := map[string]interface{}{"kind": v.Type().Name()}
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.PkgPath != "" { // unexported
				continue
			}
			out[f.Name] = astToPlain(v.Field(i))
		}
		return out

	case reflect.String:
		return v.String()
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	default:
		return nil
	}
}

// AST_json writes the AST to fout as indented JSON.
func AST_json(fout *os.File, ast AST) {
	enc := json.NewEncoder(fout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(astToPlain(reflect.ValueOf(ast))); err != nil {
		panic(err)
	}
}
