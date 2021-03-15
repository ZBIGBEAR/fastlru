/*
时间复杂度为O(1)的LRU
*/
package fastlru

import (
	"context"
	"errors"
)

const (
	MaxElem = 1000
	MinElem = 10
	DefaultElem = 100
)

var (
	EmptyErr = errors.New("[fastlru] lur is empty")
	NotFoundErr = errors.New("[fastlru] not found")
	UnknowErr = errors.New("[fastlru] unknow error")
)

type Elem struct {
	key string
	val interface{}
	next *Elem
}

type Lru interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, val interface{}) error
	Clear(ctx context.Context)
	GetAllValue(ctx context.Context) ([]interface{})
}

func NewLru(maxCount int) Lru {
	count := maxCount
	if count < MinElem {
		count = MinElem
	}else if count > MaxElem {
		count = MaxElem
	}

	emptyElem := &Elem{}
	return &lruCache{
		firstElem:    emptyElem,
		maxCount:     count,
		currentCount: 0,
		keyMap:       make(map[string]*Elem, count),
		header:       emptyElem,
		tail:         emptyElem,
	}
}

func NewDefaultLru() Lru {
	return NewLru(DefaultElem)
}

type lruCache struct {
	firstElem *Elem // 第一个节点是空，永远指向链表的头节点
	maxCount int
	currentCount int
	keyMap map[string]*Elem // 把所有的key放在map中，方便快速查找某个元素，val存储的是key节点的前面节点，方便删除key节点
	header *Elem // 头部节点
	tail *Elem // 尾节点。lru满的时候方便删除尾部节点
}

func (l *lruCache) Get(ctx context.Context, key string) (interface{}, error) {
	// 1.如果lru为空则查找失败
	if l.Empty() {
		return nil, EmptyErr
	}

	// 2.不为空
	elem, ok := l.keyMap[key]
	if !ok {
		// 没有找到
		return nil, NotFoundErr
	}

	// 3.找到了，返回对应的值，并且把当前元素移动到队首
	val := elem.next.val
	l.moveElem2Header(ctx, elem)
	return val, nil
}

func (l *lruCache) Set(ctx context.Context, key string, val interface{}) error {
	// 1.先查找
	v, err := l.Get(ctx, key)
	if err == nil {
		// 找到
		if v != val {
			l.header.next.val = val
		}
	}else if err == EmptyErr {
		l.insertFirstElem(ctx, key, val)
	}else if err == NotFoundErr {
		l.insertElem(ctx,key,val)
	}else {
		return UnknowErr
	}
	return nil
}

func (l *lruCache) Empty() bool {
	return l.header.next == nil
}

func (l *lruCache) moveElem2Header(ctx context.Context, elem *Elem) {
	if elem == l.firstElem {
		return
	}
	newFirstElem := elem.next
	if newFirstElem.next == nil {
		// 最后一个元素
		elem.next = nil
		l.tail = elem
	}else{
		//
		l.keyMap[newFirstElem.next.key] = elem
		elem.next = newFirstElem.next
	}
	// 最后一个元素插入
	oldFirstElem := l.header.next
	newFirstElem.next = oldFirstElem
	l.keyMap[oldFirstElem.key] = newFirstElem
	l.header.next = newFirstElem
	l.keyMap[newFirstElem.key] = l.header
}

func (l *lruCache) insertFirstElem(ctx context.Context, key string, val interface{}) {
	newElem := &Elem{
		key:  key,
		val:  val,
		next: nil,
	}
	l.header.next = newElem
	l.tail = newElem
	l.currentCount++
	l.keyMap[key] = l.header
}

func (l *lruCache) insertElem(ctx context.Context, key string, val interface{}) {
	if l.maxCount == l.currentCount {
		// lru已满，需要淘汰数据
		l.deleteTail(ctx)
		l.currentCount--
	}
	// 向队首插入一个元素
	newElem := &Elem{
		key:  key,
		val:  val,
		next: nil,
	}
	oldFirstElem := l.header.next
	newElem.next = oldFirstElem
	l.header.next=newElem
	l.keyMap[key] = l.header
	l.keyMap[oldFirstElem.key] = newElem
	l.currentCount++
}

// 删除末尾元素
func (l *lruCache) deleteTail(ctx context.Context) {
	deleteKey := l.tail.key
	tailPre := l.keyMap[deleteKey]
	tailPre.next = nil
	l.tail = tailPre
	delete(l.keyMap, deleteKey)
}

// 返回所有元素
func (l *lruCache) GetAllValue(ctx context.Context) ([]interface{}) {
	result := make([]interface{}, 0)
	p := l.header.next
	for p!= nil {
		result = append(result, p.val)
		p = p.next
	}
	return result
}

func (l *lruCache) Clear(ctx context.Context) {
	l.currentCount = 0
	l.keyMap = make(map[string]*Elem,l.maxCount)
	l.header = l.firstElem
	l.tail = l.firstElem
}