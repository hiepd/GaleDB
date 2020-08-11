package index

import (
	"container/list"
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
			err := si.Add(tt.args.row)
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

func TestScanIndex_Remove(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free *list.List
	}
	type args struct {
		key entity.Key
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: tt.fields.free,
			}
			if err := si.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("ScanIndex.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanIndex_Get(t *testing.T) {
	type fields struct {
		rows []*entity.Row
		free *list.List
	}
	type args struct {
		key entity.Key
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    entity.Row
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: tt.fields.free,
			}
			got, err := si.Get(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanIndex.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanIndex.Get() = %v, want %v", got, tt.want)
			}
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
		free *list.List
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIndex{
				rows: tt.fields.rows,
				free: tt.fields.free,
			}
			if got := si.Size(); got != tt.want {
				t.Errorf("ScanIndex.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanIterator_Next(t *testing.T) {
	type fields struct {
		index    *ScanIndex
		position int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIterator{
				index:    tt.fields.index,
				position: tt.fields.position,
			}
			if got := si.Next(); got != tt.want {
				t.Errorf("ScanIterator.Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanIterator_Current(t *testing.T) {
	type fields struct {
		index    *ScanIndex
		position int
	}
	tests := []struct {
		name   string
		fields fields
		want   entity.Row
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := &ScanIterator{
				index:    tt.fields.index,
				position: tt.fields.position,
			}
			if got := si.Current(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScanIterator.Current() = %v, want %v", got, tt.want)
			}
		})
	}
}
