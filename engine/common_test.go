package engine

import (
	"reflect"
	"testing"
)

func TestResource_getImplicitKeys(t *testing.T) {
	type fields struct {
		Name string
		Args map[string]interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "empty", fields: fields{
			Name: "",
			Args: map[string]interface{}{},
		}, want: nil},
		{
			name: "no_match",
			fields: fields{
				Name: "",
				Args: map[string]interface{}{
					"key": "value",
				},
			},
			want: nil,
		},
		{
			name: "match_simple",
			fields: fields{
				Name: "",
				Args: map[string]interface{}{
					"$_key": "value",
				},
			},
			want: []string{"$_key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Resource{
				Name: tt.fields.Name,
				Args: tt.fields.Args,
			}
			if got := r.getImplicitKeys(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getImplicitKeys() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
