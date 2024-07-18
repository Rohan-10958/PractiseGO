package LinkedList

type node struct {
	val  int
	next *node
}
type linkedList struct {
	head *node
}

func NewLinkedLisT() linkedList {
	return linkedList{}
}
func (ll *linkedList) AddLast(val int) {
	if ll.head == nil {
		ll.head = &node{val: val}
		return
	}
	p := ll.head
	for p.next != nil {
		p = p.next
	}
	p.next = &node{val: val}

}
func (ll *linkedList) FindVal(val int) bool {
	p := ll.head
	for p != nil {
		if p.val == val {
			return true
		}
		p = p.next
	}
	return false
}

func (ll *linkedList) FeturnAll() []int {
	p := ll.head
	allVal := make([]int, 0)
	for p != nil {
		allVal = append(allVal, p.val)
	}
	return allVal
}
