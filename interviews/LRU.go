package main

// LRU = map + 双向链表
// map：O(1) 找节点
// 双向链表：维护顺序
// 双向链表中使用虚拟的头尾节点，方便处理边界情况
// Remove(node)：从当前位置摘掉节点
// AddToFront(node)：插到 dummyHead 后面
// MoveToFront(node)：Remove + AddToFront
// RemoveTail()：删除 dummyTail 前面的真实节点

type LRUCache struct {
	capacity int
	cache    map[int]*Node
	list     *LinkedList
}

type LinkedList struct {
	dummyHead *Node // dummyHead 后面是最近使用的节点
	dummyTail *Node // dummyTail 前面是最久未使用的节点
}

type Node struct {
	key   int
	value int
	prev  *Node
	next  *Node
}

func Constructor(capacity int) LRUCache {
	dummyHead := &Node{}
	dummyTail := &Node{}

	// 初始化空双向链表：head <-> tail
	dummyHead.next = dummyTail
	dummyTail.prev = dummyHead

	return LRUCache{
		capacity: capacity,
		cache:    make(map[int]*Node),
		list: &LinkedList{
			dummyHead: dummyHead,
			dummyTail: dummyTail,
		},
	}
}

// Get 命中时，说明该 key 被访问过，要移动到最近使用位置。
func (c *LRUCache) Get(key int) int {
	node, ok := c.cache[key]
	if !ok {
		return -1
	}

	c.list.MoveToFront(node)
	return node.value
}

// Put 分两种情况：
// 1. key 已存在：更新 value，并移动到最近使用位置。
// 2. key 不存在：新建节点，放到最近使用位置；如果超容量，淘汰最久未使用节点。
func (c *LRUCache) Put(key int, value int) {
	if node, ok := c.cache[key]; ok {
		node.value = value
		c.list.MoveToFront(node)
		return
	}

	node := &Node{
		key:   key,
		value: value,
	}

	c.cache[key] = node
	c.list.AddToFront(node)

	if len(c.cache) > c.capacity {
		removed := c.list.RemoveTail()
		delete(c.cache, removed.key)
	}
}

// Remove 把真实节点从双向链表中摘掉。
// 前提：node 不是 dummyHead / dummyTail。
func (l *LinkedList) Remove(node *Node) {
	prev := node.prev
	next := node.next

	prev.next = next
	next.prev = prev

	node.prev = nil
	node.next = nil
}

// AddToFront 把节点插到 dummyHead 后面。
// 这里定义 dummyHead 后面是“最近使用”的位置。
func (l *LinkedList) AddToFront(node *Node) {
	first := l.dummyHead.next

	node.prev = l.dummyHead
	node.next = first

	l.dummyHead.next = node
	first.prev = node
}

// MoveToFront 表示某节点刚被访问或更新，移动到最近使用位置。
func (l *LinkedList) MoveToFront(node *Node) {
	l.Remove(node)
	l.AddToFront(node)
}

// RemoveTail 删除最久未使用节点，也就是 dummyTail 前面的真实节点。
func (l *LinkedList) RemoveTail() *Node {
	node := l.dummyTail.prev
	l.Remove(node)
	return node
}
