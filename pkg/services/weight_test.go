package services

import (
	"testing"
	"time"
)

func Test_calcComputedWeight(t *testing.T) {
	type args struct {
		targetPercentage         int
		totalCurrWeightOtherSvcs int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "basic 50/50 calc",
			args: args{
				targetPercentage:         50,
				totalCurrWeightOtherSvcs: 50,
			},
			want: 50,
		},
		{
			name: "basic 50/75 calc",
			args: args{
				targetPercentage:         75,
				totalCurrWeightOtherSvcs: 50,
			},
			want: 150,
		},
		{
			name: "attempt to promote always gives same answer",
			args: args{
				targetPercentage:         100,
				totalCurrWeightOtherSvcs: 100,
			},
			want: 10000,
		},
		{
			name: "basic zero scale",
			args: args{
				targetPercentage:         0,
				totalCurrWeightOtherSvcs: 100,
			},
			want: 0,
		},
		{
			name: "return target percentage when no weight",
			args: args{
				targetPercentage:         50,
				totalCurrWeightOtherSvcs: 0,
			},
			want: 50,
		},
		{
			name: "small case",
			args: args{
				targetPercentage:         50,
				totalCurrWeightOtherSvcs: 1,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcComputedWeight(tt.args.targetPercentage, tt.args.totalCurrWeightOtherSvcs); got != tt.want {
				t.Errorf("calcComputedWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calcIncrement(t *testing.T) {
	type args struct {
		duration                 time.Duration
		targetPercentage         int
		totalCurrWeight          int
		totalCurrWeightOtherSvcs int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "no change should be 0",
			args: args{
				duration:                 time.Second * 1,
				targetPercentage:         50,
				totalCurrWeight:          100,
				totalCurrWeightOtherSvcs: 50,
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "basic large case",
			args: args{
				duration:                 time.Second * 60,
				targetPercentage:         99,
				totalCurrWeight:          1000,
				totalCurrWeightOtherSvcs: 500,
			},
			want: 3267,
		},
		{
			name: "basic small case",
			args: args{
				duration:                 time.Second * 60,
				targetPercentage:         99,
				totalCurrWeight:          100,
				totalCurrWeightOtherSvcs: 50,
			},
			want: 327,
		},
		{
			name: "basic scale down case in one step",
			args: args{
				duration:                 time.Second * 4,
				targetPercentage:         50,
				totalCurrWeight:          100,
				totalCurrWeightOtherSvcs: 25,
			},
			want: 50,
		},
		{
			name: "too low of an increment should error out",
			args: args{
				duration:                 time.Hour * 10,
				targetPercentage:         50,
				totalCurrWeight:          100,
				totalCurrWeightOtherSvcs: 100,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "dramatic scale down in one step",
			args: args{
				duration:                 time.Second * 1,
				targetPercentage:         50,
				totalCurrWeight:          100,
				totalCurrWeightOtherSvcs: 1,
			},
			want: 98,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calcIncrement(tt.args.duration, tt.args.targetPercentage, tt.args.totalCurrWeight, tt.args.totalCurrWeightOtherSvcs)
			if (err != nil) != tt.wantErr {
				t.Errorf("calcIncrement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("calcIncrement() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalcWeightPercentage(t *testing.T) {
	type args struct {
		weight      int
		totalWeight int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "basic case",
			args: args{
				weight:      775,
				totalWeight: 1000,
			},
			want: 78,
		},
		{
			name: "zero weight of 100 total",
			args: args{
				weight:      0,
				totalWeight: 100,
			},
			want: 0,
		},
		{
			name: "zero weight of zero total",
			args: args{
				weight:      0,
				totalWeight: 0,
			},
			want: 0,
		},
		{
			name: ".49% rounds down",
			args: args{
				weight:      1,
				totalWeight: 201,
			},
			want: 0,
		},
		{
			name: ".50% rounds up",
			args: args{
				weight:      1,
				totalWeight: 200,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalcWeightPercentage(tt.args.weight, tt.args.totalWeight); got != tt.want {
				t.Errorf("CalcWeightPercentage() = %v, want %v", got, tt.want)
			}
		})
	}
}
