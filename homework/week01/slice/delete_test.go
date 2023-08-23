package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geektime-basic-go/homework/week01/slice/internal"
)

func TestDelete(t *testing.T) {
	testCases := []struct {
		name    string
		src     []int
		index   int
		want    []int
		wantErr error
	}{
		{
			name:    "first index",
			src:     []int{1, 2, 3, 4, 5},
			index:   0,
			want:    []int{2, 3, 4, 5},
			wantErr: nil,
		},
		{
			name:    "last index",
			src:     []int{1, 2, 3, 4, 5},
			index:   4,
			want:    []int{1, 2, 3, 4},
			wantErr: nil,
		},
		{
			name:    "other index",
			src:     []int{1, 2, 3, 4, 5},
			index:   3,
			want:    []int{1, 2, 3, 5},
			wantErr: nil,
		},
		{
			name:    "index -1",
			src:     []int{1, 2, 3, 4, 5},
			index:   -1,
			want:    []int{1, 2, 3, 4, 5},
			wantErr: internal.ErrIndexOutOfRange,
		},
		{
			name:    "index out of range",
			src:     []int{1, 2, 3, 4, 5},
			index:   100,
			want:    []int{1, 2, 3, 4, 5},
			wantErr: internal.ErrIndexOutOfRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Delete(tc.src, tc.index)
			require.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDeleteWithReduceCapacity(t *testing.T) {
	testCases := []struct {
		name string

		index int
		src   func() []int

		wantRes []int
		wantCap int

		wantErr error
	}{
		{
			name: "reduce capacity",

			index: 0,
			src: func() []int {
				res := make([]int, 0, 10)
				res = append(res, []int{1, 2, 3}...)
				return res
			},

			wantRes: []int{2, 3},
			wantCap: 5,
			wantErr: nil,
		},
		{
			name:  "without reduce capacity",
			index: 0,
			src: func() []int {
				res := make([]int, 0, 10)
				res = append(res, []int{1, 2, 3, 4, 5}...)
				return res
			},

			wantCap: 10,
			wantRes: []int{2, 3, 4, 5},
			wantErr: nil,
		},
		{
			name:  "cap less 10",
			index: 0,
			src: func() []int {
				res := make([]int, 0, 8)
				res = append(res, []int{1, 2, 3, 4, 5}...)
				return res
			},

			wantCap: 8,
			wantRes: []int{2, 3, 4, 5},
			wantErr: nil,
		},
		{
			name: "index -1",

			index: -1,
			src: func() []int {
				return []int{1, 2, 3, 4, 5}
			},

			wantRes: []int{1, 2, 3, 4, 5},
			wantErr: internal.ErrIndexOutOfRange,
		},
		{
			name: "index out of range",

			index: 100,
			src: func() []int {
				return []int{1, 2, 3, 4, 5}
			},

			wantRes: []int{1, 2, 3, 4, 5},
			wantErr: internal.ErrIndexOutOfRange,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeleteWithReduceCapacity(tc.src(), tc.index)
			require.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.wantRes, got)
		})
	}
}

func BenchmarkDelete(b *testing.B) {
	src := make([]int, 1000000)
	for i := 0; i < len(src); i++ {
		src[i] = i
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := Delete(src, 500000)
		if err != nil {
			b.Fatalf("Error: %v", err)
		}
	}
}

func BenchmarkDeleteWithReduceCapacity(b *testing.B) {
	src := make([]int, 250000, 1000000)
	for i := 0; i < len(src); i++ {
		src[i] = i
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := DeleteWithReduceCapacity(src, 125000)
		if err != nil {
			b.Fatalf("Error: %v", err)
		}
	}
}
