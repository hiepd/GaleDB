package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	type args struct {
		sql string
	}
	tests := []struct {
		name    string
		args    args
		want    Statement
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				sql: "select * from table1",
			},
			want: &Select{
				From: &From{
					TableName: "table1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.sql)
			assert.Equal(t, tt.want, got)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
