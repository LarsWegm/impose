package composeparser

import "testing"

func Test_containsOption(t *testing.T) {
	type args struct {
		comment string
		option  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"found option",
			args{"#impose:option", "option"},
			true,
		},
		{
			"no option prefix",
			args{"#noprefix:option", "option"},
			false,
		},
		{
			"multiline comment",
			args{"#line1\n#impose:option\n#line3", "option"},
			true,
		},
		{
			"inline comment",
			args{"#some comment impose:option other comment text", "option"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsOption(tt.args.comment, tt.args.option); got != tt.want {
				t.Errorf("containsOption() = %v, want %v", got, tt.want)
			}
		})
	}
}
