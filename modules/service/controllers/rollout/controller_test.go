package rollout

import "testing"

func Test_incrementFlux(t *testing.T) {
	type args struct {
		increment  int
		goalWeight int
		currWeight int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "10% curr weight should bring increment down by half",
			args: args{
				increment:  10,
				goalWeight: 100,
				currWeight: 0,
			},
			want: 5,
		},
		{
			name: "50% curr weight should leave increment",
			args: args{
				increment:  10,
				goalWeight: 100,
				currWeight: 50,
			},
			want: 10,
		},
		{
			name: "90% curr weight should increase increment by 1.25%",
			args: args{
				increment:  10,
				goalWeight: 100,
				currWeight: 90,
			},
			want: 12,
		},
		{
			name: "Large change going down should still be multiplied by 1.25%",
			args: args{
				increment:  10,
				goalWeight: 100,
				currWeight: 1000,
			},
			want: 12,
		},
		{
			name: "Small change going down should still be 50%",
			args: args{
				increment:  10,
				goalWeight: 950,
				currWeight: 1000,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := incrementFlux(tt.args.increment, tt.args.goalWeight, tt.args.currWeight); got != tt.want {
				t.Errorf("incrementFlux() = %v, want %v", got, tt.want)
			}
		})
	}
}
