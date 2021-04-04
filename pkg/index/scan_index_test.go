package index

import (
	"container/list"
	"errors"
	"reflect"
	"testing"

	"github.com/hiepd/galedb/pkg/entity"
	"github.com/stretchr/testify/assert"
)

func TestNewScanIndex(t *testing.T) {
	tests := []struct {
		name string
		want Index
	}{
		{
			name: "Success",
			want: &ScanIndex{
				rows: make([]*entity.Row, 0),
				free: list.New(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewScanIndex())
		})
	}
}

func TestScanIndex_Add(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free []int
	}
	type args struct {
		row entity.Row
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    fields
		wantKey entity.Key
		wantErr error
	}{
		{
			name: "No existing",
			fields: fields{
				rows: []*entity.Row{},
				free: nil,
			},
			args: args{
				row: entity.Row{
					Key:    0,
					Values: []entity.Value{"val1"},
				},
			},
			want: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
				},
				free: nil,
			},
			wantKey: 1,
			wantErr: nil,
		},
		{
			name: "Some existing - No Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: nil,
			},
			args: args{
				row: entity.Row{
					Key:    0,
					Values: []entity.Value{"val3"},
				},
			},
			want: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					{
						Key:    3,
						Values: []entity.Value{"val3"},
					},
				},
				free: nil,
			},
			wantKey: 3,
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: []int{0},
			},
			args: args{
				row: entity.Row{
					Key:    0,
					Values: []entity.Value{"val3"},
				},
			},
			want: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val3"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: nil,
			},
			wantKey: 1,
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes 2",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free: []int{0, 2},
			},
			args: args{
				row: entity.Row{
					Key:    0,
					Values: []entity.Value{"val3"},
				},
			},
			want: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val3"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free: []int{2},
			},
			wantKey: 1,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: list.New(),
			}
			for _, i := range tt.fields.free {
				si.free.PushBack(i)
			}
			key, err := si.Add(tt.args.row)
			assert.Equal(t, tt.wantErr, err)
			assert.ElementsMatch(t, tt.want.rows, si.rows)
			assert.Equal(t, len(tt.want.free), si.free.Len())
			assert.Equal(t, tt.wantKey, key)
			e := si.free.Front()
			for _, i := range tt.want.free {
				assert.Equal(t, i, e.Value)
				e = e.Next()
			}
		})
	}
}

func TestScanIndex_Remove(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free []int
	}
	type args struct {
		key entity.Key
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    fields
		wantErr error
	}{
		{
			name: "No existing",
			fields: fields{
				rows: []*entity.Row{},
				free: nil,
			},
			args: args{
				key: 1,
			},
			want: fields{
				rows: []*entity.Row{},
				free: nil,
			},
			wantErr: errors.New("invalid key"),
		},
		{
			name: "Some existing - No Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: nil,
			},
			args: args{
				key: 1,
			},
			want: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: []int{0},
			},
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: []int{0},
			},
			args: args{
				key: 2,
			},
			want: fields{
				rows: []*entity.Row{nil, nil},
				free: []int{0, 1},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: list.New(),
			}
			for _, i := range tt.fields.free {
				si.free.PushBack(i)
			}
			err := si.Remove(tt.args.key)
			assert.Equal(t, tt.wantErr, err)
			assert.ElementsMatch(t, tt.want.rows, si.rows)
			assert.Equal(t, len(tt.want.free), si.free.Len())
			e := si.free.Front()
			for _, i := range tt.want.free {
				assert.Equal(t, i, e.Value)
				e = e.Next()
			}
		})
	}
}

func TestScanIndex_Get(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free []int
	}
	type args struct {
		key entity.Key
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entity.Row
		wantErr error
	}{
		{
			name: "No existing",
			fields: fields{
				rows: []*entity.Row{},
				free: nil,
			},
			args: args{
				key: 1,
			},
			want:    entity.Row{},
			wantErr: errors.New("invalid key"),
		},
		{
			name: "Some existing - No Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: nil,
			},
			args: args{
				key: 1,
			},
			want: entity.Row{
				Key:    1,
				Values: []entity.Value{"val1"},
			},
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: []int{0},
			},
			args: args{
				key: 2,
			},
			want: entity.Row{
				Key:    2,
				Values: []entity.Value{"val2"},
			},
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes 2",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free: []int{0, 2},
			},
			args: args{
				key: 2,
			},
			want: entity.Row{
				Key:    2,
				Values: []entity.Value{"val2"},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: list.New(),
			}
			for _, i := range tt.fields.free {
				si.free.PushBack(i)
			}
			row, err := si.Get(tt.args.key)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, row)
		})
	}
}

func TestScanIndex_Iterator(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free *list.List
	}
	tests := []struct {
		name   string
		fields fields
		want   Iterator
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: tt.fields.free,
			}
			if got := si.Iterator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanIndex.Iterator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanIndex_Size(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free []int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "No existing",
			fields: fields{
				rows: []*entity.Row{},
				free: nil,
			},
			want: 0,
		},
		{
			name: "Some existing - No Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: nil,
			},
			want: 2,
		},
		{
			name: "Some existing - Some Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free: []int{0},
			},
			want: 1,
		},
		{
			name: "Some existing - Some Free Indexes 2",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free: []int{0, 2},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: list.New(),
			}
			for _, i := range tt.fields.free {
				si.free.PushBack(i)
			}
			size := si.Size()
			assert.Equal(t, tt.want, size)
		})
	}
}

func TestScanIterator_Next(t *testing.T) {
	type fields struct {
		rows     []*entity.Row
		free     []int
		position int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "No existing",
			fields: fields{
				rows:     []*entity.Row{},
				free:     nil,
				position: -1,
			},
			want: false,
		},
		{
			name: "Some existing - No Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free:     nil,
				position: -1,
			},
			want: true,
		},
		{
			name: "Some existing - No Free Indexes 1",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free:     nil,
				position: 1,
			},
			want: false,
		},
		{
			name: "Some existing - Some Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free:     []int{0},
				position: -1,
			},
			want: true,
		},
		{
			name: "Some existing - Some Free Indexes 2",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free:     []int{0, 2},
				position: -1,
			},
			want: true,
		},
		{
			name: "Some existing - Some Free Indexes 3",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free:     []int{0, 2},
				position: -1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sidx := &ScanIndex{
				rows: tt.fields.rows,
				free: list.New(),
			}
			for _, i := range tt.fields.free {
				sidx.free.PushBack(i)
			}
			si := &ScanIterator{
				index:    sidx,
				position: tt.fields.position,
			}
			assert.Equal(t, tt.want, si.Next())
		})
	}
}

func TestScanIterator_Current(t *testing.T) {
	type fields struct {
		rows     []*entity.Row
		free     []int
		position int
	}
	tests := []struct {
		name    string
		fields  fields
		want    entity.Row
		wantErr error
	}{
		{
			name: "No existing",
			fields: fields{
				rows:     []*entity.Row{},
				free:     nil,
				position: -1,
			},
			want:    entity.Row{},
			wantErr: errors.New("invalid cursor"),
		},
		{
			name: "No existing 1",
			fields: fields{
				rows:     []*entity.Row{},
				free:     nil,
				position: 0,
			},
			want:    entity.Row{},
			wantErr: errors.New("invalid cursor"),
		},
		{
			name: "Some existing - No Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free:     nil,
				position: 0,
			},
			want: entity.Row{
				Key:    1,
				Values: []entity.Value{"val1"},
			},
			wantErr: nil,
		},
		{
			name: "Some existing - No Free Indexes 1",
			fields: fields{
				rows: []*entity.Row{
					{
						Key:    1,
						Values: []entity.Value{"val1"},
					},
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free:     nil,
				position: 1,
			},
			want: entity.Row{
				Key:    2,
				Values: []entity.Value{"val2"},
			},
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
				},
				free:     []int{0},
				position: 0,
			},
			want:    entity.Row{},
			wantErr: errors.New("invalid cursor"),
		},
		{
			name: "Some existing - Some Free Indexes 2",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free:     []int{0, 2},
				position: 1,
			},
			want: entity.Row{
				Key:    2,
				Values: []entity.Value{"val2"},
			},
			wantErr: nil,
		},
		{
			name: "Some existing - Some Free Indexes 3",
			fields: fields{
				rows: []*entity.Row{
					nil,
					{
						Key:    2,
						Values: []entity.Value{"val2"},
					},
					nil,
				},
				free:     []int{0, 2},
				position: 2,
			},
			want:    entity.Row{},
			wantErr: errors.New("invalid cursor"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sidx := &ScanIndex{
				rows: tt.fields.rows,
				free: list.New(),
			}
			for _, i := range tt.fields.free {
				sidx.free.PushBack(i)
			}
			si := &ScanIterator{
				index:    sidx,
				position: tt.fields.position,
			}
			row, err := si.Current()
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, row)
		})
	}
}
