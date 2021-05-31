package skiplist

type Interface interface {
	Less(other interface{}) bool
}

type skiplistLevel struct {
	forward *Element
	span    int
}

type Element struct {
	Value    Interface
	backward *Element
	level    []*skiplistLevel
}

// Next returns the next skiplist element or nil.
func (e *Element) Next() *Element {
	return e.level[0].forward
}

// Prev returns the previous skiplist element of nil.
func (e *Element) Prev() *Element {
	return e.backward
}

// newElement returns an initialized element.
func newElement(level int, v Interface) *Element {
	slLevels := make([]*skiplistLevel, level)
	for i := 0; i < level; i++ {
		slLevels[i] = new(skiplistLevel)
	}

	return &Element{
		Value:    v,
		backward: nil,
		level:    slLevels,
	}
}
