package lru

import "container/list"

//TODO
/*
	可改进的方案:lru-k
*/

// Cache lru实现
type Cache struct {
	maxBytes  int64                         //允许使用的最大内存
	nbytes    int64                         //当前已经使用的内存
	ll        *list.List                    //双向链表
	cache     map[string]*list.Element      //哈希表
	OnEvicted func(key string, value Value) //记录被删除时的回调函数
}

// entry 在双向链表中保存键值对
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int //返回值所占用的内存大小
}

// Get 查找，通过key获取值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 删除,淘汰掉最近最少访问的节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 新增
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//maxBytes为0表示不对内存大小做限制
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// New 初始化LRU列表
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}
