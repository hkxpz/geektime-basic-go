package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			wantErr: ErrIndexOutOfRange,
		},
		{
			name:    "index out of range",
			src:     []int{1, 2, 3, 4, 5},
			index:   100,
			wantErr: ErrIndexOutOfRange,
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
