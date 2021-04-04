package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Lexer_Lex(t *testing.T) {
	type fields struct {
		input []byte
		pos   int
		err   error
	}
	type args struct {
		lval *yySymType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int
	}{
		{
			name: "name",
			fields: fields{
				input: []byte("hello"),
			},
			args: args{
				lval: &yySymType{},
			},
			want: []int{NAME},
		},
		{
			name: "select",
			fields: fields{
				input: []byte("select"),
			},
			args: args{
				lval: &yySymType{},
			},
			want: []int{SELECT},
		},
		{
			name: "from",
			fields: fields{
				input: []byte("from"),
			},
			args: args{
				lval: &yySymType{},
			},
			want: []int{FROM},
		},
		{
			name: "statement 1",
			fields: fields{
				input: []byte("select from"),
			},
			args: args{
				lval: &yySymType{},
			},
			want: []int{SELECT, FROM},
		},
		{
			name: "statement 2",
			fields: fields{
				input: []byte("select from;"),
			},
			args: args{
				lval: &yySymType{},
			},
			want: []int{SELECT, FROM},
		},
		{
			name: "statement 2",
			fields: fields{
				input: []byte("select * from table1;"),
			},
			args: args{
				lval: &yySymType{},
			},
			want: []int{SELECT, ASTERISK, FROM, NAME},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Lexer{
				Input: tt.fields.input,
				Pos:   tt.fields.pos,
				Err:   tt.fields.err,
			}
			got := make([]int, 0)
			for res := l.Lex(tt.args.lval); res != 0; res = l.Lex(tt.args.lval) {
				got = append(got, res)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
