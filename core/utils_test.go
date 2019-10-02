package core

import (
	"github.com/ofux/deluge/dsl/parser"
	"testing"
)

func TestSPrintParserErrors(t *testing.T) {
	type args struct {
		errors []parser.ParseError
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "No error",
			args: args{errors: []parser.ParseError{}},
			want: "",
		}, {
			name: "One error",
			args: args{errors: []parser.ParseError{
				{
					Message: "This is an error",
					Line:    10,
					Column:  5,
				},
			}},
			want: "Syntax error:\n\tThis is an error (line 10, col 5)\n",
		}, {
			name: "Multiple errors",
			args: args{errors: []parser.ParseError{
				{
					Message: "This is an error",
					Line:    10,
					Column:  5,
				}, {
					Message: "This is another error",
					Line:    0,
					Column:  0,
				}, {
					Message: "This is one more error",
					Line:    1,
					Column:  5020,
				},
			}},
			want: "Syntax error:\n\tThis is an error (line 10, col 5)\n\tThis is another error (line 0, col 0)\n\tThis is one more error (line 1, col 5020)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SPrintParserErrors(tt.args.errors); got != tt.want {
				t.Errorf("SPrintParserErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}
