package michi

import (
	"testing"
)

func Test_methodAndPath(t *testing.T) {
	type args struct {
		pattern string
	}
	tests := []struct {
		name       string
		args       args
		wantMethod string
		wantPath   string
	}{
		{
			name: "/a",
			args: args{
				pattern: "/a",
			},
			wantMethod: "",
			wantPath:   "/a",
		},
		{
			name: "POST /a",
			args: args{
				pattern: "POST /a",
			},
			wantMethod: "POST",
			wantPath:   "/a",
		},
		{
			name: "POST example.com/a",
			args: args{
				pattern: "POST example.com/a",
			},
			wantMethod: "POST",
			wantPath:   "example.com/a",
		},
		{
			name: "POST /a/{b}",
			args: args{
				pattern: "POST /a/{b}",
			},
			wantMethod: "POST",
			wantPath:   "/a/{b}",
		},
		// Allow multiple spaces between method and path.
		{
			name: "POST  /a",
			args: args{
				pattern: "POST  /a",
			},
			wantMethod: "POST",
			wantPath:   "/a",
		},
		{
			name: "POST  example.com/a",
			args: args{
				pattern: "POST  example.com/a",
			},
			wantMethod: "POST",
			wantPath:   "example.com/a",
		},
		{
			name: "POST  /a/{b}",
			args: args{
				pattern: "POST  /a/{b}",
			},
			wantMethod: "POST",
			wantPath:   "/a/{b}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMethod, gotPath := methodAndPath(tt.args.pattern)
			if gotMethod != tt.wantMethod {
				t.Errorf("method got = %v, want %v", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wantPath {
				t.Errorf("path got = %v, want %v", gotPath, tt.wantPath)
			}
		})
	}
}
