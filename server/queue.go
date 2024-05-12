package main

// Queue represents a FIFO queue
type Queue struct {
	items []interface{}
}

// Enqueue adds an item to the end of the queue
func (q *Queue) Enqueue(item interface{}) {
	q.items = append(q.items, item)
}

// Dequeue removes and returns the item at the front of the queue
func (q *Queue) Dequeue() interface{} {
	if len(q.items) == 0 {
		return nil // Return nil if queue is empty
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item
}

// IsEmpty returns true if the queue is empty
func (q *Queue) IsEmpty() bool {
	return len(q.items) == 0
}

// Size returns the number of items in the queue
func (q *Queue) Size() int {
	return len(q.items)
}
