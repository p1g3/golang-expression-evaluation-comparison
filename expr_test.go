package main

import (
	"testing"

	"github.com/antonmedv/expr"
)

func Benchmark_expr(b *testing.B) {
	params := createParams()

	program, err := expr.Compile(example, expr.Env(params))
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expr.Run(program, params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_expr_startswith(b *testing.B) {
	params := map[string]interface{}{
		"name":  "/groups/foo/bar",
		"group": "foo",
	}

	program, err := expr.Compile(`name startsWith "/groups/" + group`, expr.Env(params))
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expr.Run(program, params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_expr_funccall(b *testing.B) {
	params := map[string]interface{}{
		"hello": func(str string) string { return "hello " + str },
	}

	program, err := expr.Compile(`hello("world")`)
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expr.Run(program, params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if out.(string) != "hello world" {
		b.Fail()
	}
}

func Benchmark_expr_struct(b *testing.B) {
	request := Request{Method: "GET", Url: "http://www.google.com", Header: map[string]string{"Content-Type": "application/json"}}

	env := map[string]interface{}{
		"request": request,
	}
	program, err := expr.Compile(`request.Method == "GET" && request.Url == "http://www.google.com"`)
	if err != nil {
		b.Fatal(err)
	}

	var out interface{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expr.Run(program, env)
	}
	b.StopTimer()
	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}
