package httphandler

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_addTimestampToThisData(t *testing.T) {
	type args struct {
		paramMap map[string]string
		path     string
	}
	tests := []struct {
		name  string
		args  args
		check func(t *testing.T, paramMap map[string]string)
	}{
		{
			name: "Add timestamp with empty path",
			args: args{
				paramMap: map[string]string{
					"field1": "value1",
					"field2": "value2",
				},
				path: "",
			},
			check: func(t *testing.T, paramMap map[string]string) {
				timestamp := paramMap["timestamp"]
				assert.NotEmpty(t, timestamp, "timestamp should not be empty")

				for key, value := range paramMap {
					if key == "timestamp" {
						assert.Equal(t, timestamp, value, "timestamps should match")
					}
					if key == "field1_timestamp" || key == "field2_timestamp" {
						assert.Equal(t, timestamp, value, "timestamps should match")
					}
				}
			},
		},
		{
			name: "Add timestamp with non-empty path",
			args: args{
				paramMap: map[string]string{
					"field1": "value1",
					"field2": "value2",
				},
				path: "nested",
			},
			check: func(t *testing.T, paramMap map[string]string) {
				timestamp := paramMap["timestamp"]
				assert.NotEmpty(t, timestamp, "timestamp should not be empty")

				nestedTimestamp := paramMap["nested_timestamp"]
				assert.Empty(t, nestedTimestamp, "nested_timestamp should not be empty")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addTimestampToThisData(tt.args.paramMap, tt.args.path)
			tt.check(t, tt.args.paramMap)
		})
	}
}

func Test_collectParams(t *testing.T) {
	tests := []struct {
		name   string
		params map[string][]string
		want   map[string]string
	}{
		{
			name: "Single parameter with one value",
			params: map[string][]string{
				"param1": {"value1"},
			},
			want: map[string]string{
				"param1": "value1",
			},
		},
		{
			name: "Single parameter with multiple values",
			params: map[string][]string{
				"param1": {"value1", "value2"},
			},
			want: map[string]string{
				"param1": "value1",
			},
		},
		{
			name: "Multiple parameters with single values",
			params: map[string][]string{
				"param1": {"value1"},
				"param2": {"value2"},
			},
			want: map[string]string{
				"param1": "value1",
				"param2": "value2",
			},
		},
		{
			name: "Empty parameter value",
			params: map[string][]string{
				"param1": {},
			},
			want: map[string]string{},
		},
		{
			name: "Mixed parameters with empty and non-empty values",
			params: map[string][]string{
				"param1": {"value1"},
				"param2": {},
			},
			want: map[string]string{
				"param1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectParams(tt.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
