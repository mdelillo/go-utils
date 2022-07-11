package data_test

import (
	"github.com/mdelillo/go-utils/data"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestHeap(t *testing.T) {
	spec.Run(t, "Heap", testHeap, spec.Report(report.Terminal{}))
}

func testHeap(t *testing.T, context spec.G, it spec.S) {
	it("works as a min heap", func() {
		minHeap := data.NewMinHeap[int]()

		assert.True(t, minHeap.IsEmpty())
		assert.Equal(t, 0, minHeap.Size())

		minHeap.Push(5)

		assert.False(t, minHeap.IsEmpty())
		assert.Equal(t, 1, minHeap.Size())
		assert.Equal(t, 5, minHeap.Peek())

		minHeap.Push(2, 3, 4, 6)

		assert.Equal(t, 5, minHeap.Size())
		assert.Equal(t, 2, minHeap.Peek())

		assert.Equal(t, 2, minHeap.Pop())

		assert.Equal(t, 4, minHeap.Size())
		assert.Equal(t, 3, minHeap.Peek())

		minHeap.Pop()
		minHeap.Pop()
		minHeap.Pop()
		minHeap.Pop()

		assert.Equal(t, 0, minHeap.Peek())
		assert.Equal(t, 0, minHeap.Pop())
	})

	it("works as a max heap", func() {
		minHeap := data.NewMaxHeap[int]()

		assert.True(t, minHeap.IsEmpty())
		assert.Equal(t, 0, minHeap.Size())

		minHeap.Push(5)

		assert.False(t, minHeap.IsEmpty())
		assert.Equal(t, 1, minHeap.Size())
		assert.Equal(t, 5, minHeap.Peek())

		minHeap.Push(4, 6, 7, 8)

		assert.Equal(t, 5, minHeap.Size())
		assert.Equal(t, 8, minHeap.Peek())

		assert.Equal(t, 8, minHeap.Pop())

		assert.Equal(t, 4, minHeap.Size())
		assert.Equal(t, 7, minHeap.Peek())

		minHeap.Pop()
		minHeap.Pop()
		minHeap.Pop()
		minHeap.Pop()

		assert.Equal(t, 0, minHeap.Peek())
		assert.Equal(t, 0, minHeap.Pop())
	})

	it("properly initializes the heap", func() {
		for _, digits := range [][]int{
			{2, 8, 4, 1, 6, 7, 3, 5, 9},
			{4, 8, 2, 5, 1, 9, 6, 7, 3},
			{9, 7, 4, 1, 8, 3, 5, 2, 6},
			{7, 2, 8, 6, 4, 1, 9, 3, 5},
			{5, 7, 6, 4, 1, 8, 2, 3, 9},
		} {
			minHeap := data.NewMinHeap[int](digits...)
			assert.Equal(t, 1, minHeap.Pop())
			assert.Equal(t, 2, minHeap.Pop())
			assert.Equal(t, 3, minHeap.Pop())
			assert.Equal(t, 4, minHeap.Pop())
			assert.Equal(t, 5, minHeap.Pop())
			assert.Equal(t, 6, minHeap.Pop())
			assert.Equal(t, 7, minHeap.Pop())
			assert.Equal(t, 8, minHeap.Pop())
			assert.Equal(t, 9, minHeap.Pop())
		}
	})

	it("can use a custom comparator", func() {
		numbers := map[string]int{
			"one":   1,
			"two":   2,
			"three": 3,
			"four":  4,
			"five":  5,
			"six":   6,
		}

		minHeap := data.NewHeap[string](func(a, b string) bool {
			return numbers[a] < numbers[b]
		})

		for str := range numbers {
			minHeap.Push(str)
		}

		assert.Equal(t, "one", minHeap.Pop())
		assert.Equal(t, "two", minHeap.Pop())
		assert.Equal(t, "three", minHeap.Pop())
		assert.Equal(t, "four", minHeap.Pop())
		assert.Equal(t, "five", minHeap.Pop())
		assert.Equal(t, "six", minHeap.Pop())
	})

	it("nicely formats the heap as a tree", func() {
		minHeap := data.NewMinHeap[int](1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13)
		assert.Equal(t, strings.Trim(`
               1
       2               3
   4       5       6       7
 8   9  10  11  12  13
`, "\n"), minHeap.String(),
		)
	})
}
