package refiller

import (
	"reflect"
	"testing"
)

func TestColumns(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			args: args{
				v: struct {
					ID   int    `column:"id"`
					Name string `column:"name"`
				}{},
			},
			want: []string{"id", "name"},
		},
		{
			args: args{
				v: &(struct {
					ID   int    `column:"id"`
					Name string `column:"name"`
				}{}),
			},
			want: []string{"id", "name"},
		},
		{
			args: args{
				v: ([]*struct {
					ID   int    `column:"id"`
					Name string `column:"name"`
				}{}),
			},
			want: []string{"id", "name"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Columns(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Columns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValues(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			args: args{
				v: struct {
					ID   int    `column:"id"`
					Name string `column:"name"`
				}{
					ID:   1,
					Name: "Marty McFly",
				},
			},
			want: []interface{}{1, "Marty McFly"},
		},
		{
			args: args{
				v: &(struct {
					ID   int    `column:"id"`
					Name string `column:"name"`
				}{
					ID:   1,
					Name: "Marty McFly",
				}),
			},
			want: []interface{}{1, "Marty McFly"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Values(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Values() = %v, want %v", got, tt.want)
			}
		})
	}
}
