package data

import (
	"fmt"
	"golang.org/x/exp/constraints"
	"math"
	"strings"
)

type Heap[T any] struct {
	elements   []T
	comparator func(a, b T) bool
}

func NewHeap[T any](comparator func(a, b T) bool, elements ...T) Heap[T] {
	heap := Heap[T]{
		elements:   elements,
		comparator: comparator,
	}
	heap.init()
	return heap
}

func NewMinHeap[T constraints.Ordered](elements ...T) Heap[T] {
	return NewHeap(func(a, b T) bool { return a < b }, elements...)
}

func NewMaxHeap[T constraints.Ordered](elements ...T) Heap[T] {
	return NewHeap(func(a, b T) bool { return b < a }, elements...)
}

func (h *Heap[T]) Push(elements ...T) {
	for _, element := range elements {
		h.elements = append(h.elements, element)
		h.up(h.Size() - 1)
	}
}

func (h *Heap[T]) Peek() T {
	if h.IsEmpty() {
		return *new(T)
	}

	return h.elements[0]
}

func (h *Heap[T]) Pop() T {
	if h.IsEmpty() {
		return *new(T)
	}

	n := h.Size() - 1
	if n > 0 {
		h.swap(0, n)
		h.down(0, n)
	}
	element := h.elements[n]
	h.elements = h.elements[0:n]
	return element
}

func (h *Heap[T]) Size() int {
	return len(h.elements)
}

func (h *Heap[T]) IsEmpty() bool {
	return h.Size() == 0
}

func (h *Heap[T]) String() string {
	longestElementLength := 0
	for _, element := range h.elements {
		length := len(fmt.Sprintf("%v", element))
		if length > longestElementLength {
			longestElementLength = length
		}
	}

	levels := int(math.Log2(float64(h.Size()))) + 1

	levelPadding := 0
	elementPadding := longestElementLength
	for level := 1; level < levels; level++ {
		levelPadding = elementPadding
		elementPadding = elementPadding*2 + longestElementLength
	}

	result := ""
	for level := 1; level <= levels; level++ {
		start := int(math.Pow(2, float64(level-1)) - 1)
		end := int(math.Pow(2, float64(level))) - 1

		result += strings.Repeat(" ", levelPadding)
		for i := start; i < end && i < h.Size(); i++ {
			elementLength := len(fmt.Sprintf("%v", h.elements[i]))
			leftPadding := int(math.Ceil(float64(longestElementLength-elementLength) / 2))
			rightPadding := int(math.Floor(float64(longestElementLength-elementLength) / 2))
			result += fmt.Sprintf(
				"%s%v%s%s",
				strings.Repeat(" ", leftPadding),
				h.elements[i],
				strings.Repeat(" ", rightPadding),
				strings.Repeat(" ", elementPadding),
			)
		}
		result = strings.TrimRight(result, " ") + "\n"

		elementPadding = levelPadding
		levelPadding = (levelPadding - longestElementLength) / 2
	}
	return strings.TrimRight(result, "\n")
}

func (h *Heap[T]) init() {
	for i := len(h.elements)/2 - 1; i >= 0; i-- {
		h.down(i, len(h.elements))
	}
}

func (h *Heap[T]) up(i int) {
	for {
		parent := (i - 1) / 2

		if parent == i || !h.comparator(h.elements[i], h.elements[parent]) {
			break
		}
		h.swap(i, parent)
		i = parent
	}
}

func (h *Heap[T]) down(parent, n int) {
	for {
		leftChild := parent*2 + 1
		if leftChild >= n || leftChild < 0 {
			break
		}
		child := leftChild

		rightChild := parent*2 + 2
		if rightChild < n && h.comparator(h.elements[rightChild], h.elements[leftChild]) {
			child = rightChild
		}

		if !h.comparator(h.elements[child], h.elements[parent]) {
			break
		}

		h.swap(parent, child)
		parent = child
	}
}

func (h *Heap[T]) swap(i, j int) {
	h.elements[i], h.elements[j] = h.elements[j], h.elements[i]
}
