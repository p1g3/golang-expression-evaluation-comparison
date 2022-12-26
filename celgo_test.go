package main

import (
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	"github.com/google/cel-go/common/types"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

func Benchmark_celgo(b *testing.B) {
	params := createParams()

	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewIdent("Origin", decls.String, nil),
			decls.NewIdent("Country", decls.String, nil),
			decls.NewIdent("Value", decls.Int, nil),
			decls.NewIdent("Adults", decls.Int, nil),
		),
	)
	if err != nil {
		b.Fatal(err)
	}

	parsed, issues := env.Parse(example)
	if issues != nil && issues.Err() != nil {
		b.Fatalf("parse error: %s", issues.Err())
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		b.Fatalf("type-check error: %s", issues.Err())
	}
	prg, err := env.Program(checked)
	if err != nil {
		b.Fatalf("program construction error: %s", err)
	}

	var out ref.Val

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, _, err = prg.Eval(params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.Value().(bool) {
		b.Fail()
	}
}

func Benchmark_celgo_startswith(b *testing.B) {
	params := map[string]interface{}{
		"name":  "/groups/foo/bar",
		"group": "foo",
	}

	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewIdent("name", decls.String, nil),
			decls.NewIdent("group", decls.String, nil),
		),
	)
	if err != nil {
		b.Fatal(err)
	}

	parsed, issues := env.Parse(`name.startsWith("/groups/" + group)`)
	if issues != nil && issues.Err() != nil {
		b.Fatalf("parse error: %s", issues.Err())
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		b.Fatalf("type-check error: %s", issues.Err())
	}
	prg, err := env.Program(checked)
	if err != nil {
		b.Fatalf("program construction error: %s", err)
	}

	var out ref.Val

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, _, err = prg.Eval(params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.Value().(bool) {
		b.Fail()
	}
}

func Benchmark_celgo_funccall(b *testing.B) {
	env, err := cel.NewEnv(cel.Declarations(
		decls.NewFunction("hello", decls.NewOverload("hello_string", []*exprpb.Type{decls.String}, decls.String)),
	))
	if err != nil {
		b.Fatal(err)
	}

	ast, issue := env.Compile(`hello("world")`)
	if issue.Err() != nil {
		b.Fatal(issue.Err())
	}

	prg, err := env.Program(ast, cel.Functions(&functions.Overload{
		Operator: "hello",
		Unary: func(value ref.Val) ref.Val {
			return types.String("hello " + value.Value().(string))
		},
	}))
	if err != nil {
		b.Fatal(err)
	}

	var out ref.Val

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, _, err = prg.Eval(map[string]interface{}{})
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}

	if out.Value().(string) != "hello world" {
		b.Fail()
	}
}

type Request struct {
	ref.Val
	Method string
	Url    string
	Header map[string]string
}

func (r Request) Get(index ref.Val) ref.Val {
	switch index.(types.String).Value() {
	case "method":
		return types.String(r.Method)
	case "url":
		return types.String(r.Url)
	case "headers":
		return types.NewStringStringMap(nil, r.Header)
	}

	return nil
}

type CustomTypeProvider struct {
	ref.TypeRegistry
	typeMap map[string]map[string]*exprpb.Type
}

func (c *CustomTypeProvider) FindType(typeName string) (*exprpb.Type, bool) {
	if _, ok := c.typeMap[typeName]; ok {
		return decls.NewObjectType(typeName), true
	}
	return c.TypeRegistry.FindType(typeName)
}

func (c *CustomTypeProvider) FindFieldType(messageName, fieldName string) (*ref.FieldType, bool) {
	refs, ok := c.typeMap[messageName]
	if !ok {
		return c.TypeRegistry.FindFieldType(messageName, fieldName)
	}
	return &ref.FieldType{
		Type: refs[fieldName],
	}, true
}

func NewCustomTypeProvider() *CustomTypeProvider {
	if r, err := types.NewRegistry(); err != nil {
		panic(err)
	} else {
		return &CustomTypeProvider{
			TypeRegistry: r,
			typeMap:      make(map[string]map[string]*exprpb.Type),
		}
	}
}

func Benchmark_celgo_struct(b *testing.B) {
	c := NewCustomTypeProvider()
	c.typeMap["request"] = map[string]*exprpb.Type{
		"method":  decls.String,
		"url":     decls.String,
		"headers": decls.NewMapType(decls.String, decls.String),
	}

	request := Request{Method: "GET", Url: "http://www.google.com", Header: map[string]string{"Content-Type": "application/json"}}

	env, err := cel.NewEnv(cel.CustomTypeProvider(c))
	if err != nil {
		b.Fatal(err)
	}
	ast, issue := env.Compile(`request.method == "GET" && request.url.startsWith("http://") && request.headers["Content-Type"] == "application/json"`)
	if issue.Err() != nil {
		b.Fatal(err)
	}

	prg, err := env.Program(ast)
	if err != nil {
		b.Fatal(err)
	}
	var out ref.Val
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, _, err = prg.Eval(map[string]interface{}{"request": request})
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}

	if !out.Value().(bool) {
		b.Fail()
	}
}
