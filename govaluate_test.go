package main

import (
	"testing"

	"github.com/Knetic/govaluate"
)

func Benchmark_govaluate(b *testing.B) {
	params := createParams()

	expression, err := govaluate.NewEvaluableExpression(example)

	if err != nil {
		b.Fatal(err)
	}

	var out interface{}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expression.Evaluate(params)
	}
	b.StopTimer()

	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}

func Benchmark_govaluate_funccall(b *testing.B) {
	functions := map[string]govaluate.ExpressionFunction{
		"hello": func(args ...interface{}) (interface{}, error) {
			return "hello " + args[0].(string), nil
		},
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(`hello("world")`, functions)
	if err != nil {
		b.Fatal(err)
	}
	var out interface{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expression.Evaluate(nil)
	}
	b.StopTimer()
	if err != nil {
		b.Fatal(err)
	}
	if out.(string) != "hello world" {
		b.Fail()
	}
}

func Benchmark_govaluate_struct(b *testing.B) {
	request := Request{Method: "GET", Url: "http://www.google.com", Header: map[string]string{"Content-Type": "application/json"}}
	expression, err := govaluate.NewEvaluableExpression(`request.Method == "GET" && request.Url == "http://www.google.com"`)
	if err != nil {
		b.Fatal(err)
	}
	var out interface{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		out, err = expression.Evaluate(map[string]interface{}{"request": request})
	}
	b.StopTimer()
	if err != nil {
		b.Fatal(err)
	}
	if !out.(bool) {
		b.Fail()
	}
}
