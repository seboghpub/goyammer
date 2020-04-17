package internal

import "testing"

func Test_elipseMe(t *testing.T) {
	type args struct {
		s      string
		length int
		pad    bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "too long",
			args: args{s: "hallo", length: 2, pad: false},
			want: "h…",
		},
		{
			name: "too short with pad",
			args: args{s: "hallo", length: 10, pad: true},
			want: "hallo     ",
		},
		{
			name: "too short without pad",
			args: args{s: "hallo", length: 10, pad: false},
			want: "hallo",
		},
		{
			name: "exact",
			args: args{s: "hallo", length: 5, pad: false},
			want: "hallo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ElipseMe(tt.args.s, tt.args.length, tt.args.pad); got != tt.want {
				t.Errorf("elipseMe() = %v, want %v", got, tt.want)
			}
		})
	}
}
