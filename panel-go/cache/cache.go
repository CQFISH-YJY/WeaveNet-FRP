package cache

import (
	"sync"
	"time"
)

// Cache 内存缓存，替代 Redis 功能，进程重启数据会重置（会话有 DB 兜底）。
type Cache struct {
	mu sync.RWMutex
	m  map[string]entry
}

type entry struct {
	value any
	exp   time.Time
}

var C = &Cache{m: make(map[string]entry)}

// Set 写入带 TTL 的缓存。
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = entry{value: value, exp: time.Now().Add(ttl)}
}

// Get 读取缓存。
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	e, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.exp) {
		c.Del(key)
		return nil, false
	}
	return e.value, true
}

// Del 删除缓存。
func (c *Cache) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, key)
}

// DelPrefix 按前缀删除。
func (c *Cache) DelPrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.m {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.m, k)
		}
	}
}

// KeysPrefix 返回匹配前缀的 key 列表。
func (c *Cache) KeysPrefix(prefix string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var keys []string
	for k := range c.m {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys
}

// IncrBy 累加计数器并返回新值（带 TTL）。
func (c *Cache) IncrBy(key string, delta int64, ttl time.Duration) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var v int64
	if e, ok := c.m[key]; ok && time.Now().Before(e.exp) {
		if n, ok := e.value.(int64); ok {
			v = n
		}
	}
	v += delta
	c.m[key] = entry{value: v, exp: time.Now().Add(ttl)}
	return v
}

// GetInt64 读取整数缓存。
func (c *Cache) GetInt64(key string) (int64, bool) {
	v, ok := c.Get(key)
	if !ok {
		return 0, false
	}
	n, ok := v.(int64)
	return n, ok
}

// GetAndDel 读取并删除（用于流量聚合落库）。
func (c *Cache) GetAndDel(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.m[key]
	if !ok {
		return nil, false
	}
	delete(c.m, key)
	return e.value, true
}

// TrafficEntry 隧道当日流量。
type TrafficEntry struct {
	In  int64 `json:"in"`
	Out int64 `json:"out"`
}

// AddTraffic 累加隧道当日流量（按 date:tunnel_id）。
func AddTraffic(date string, tunnelID uint, in, out int64) {
	key := "traffic:" + date + ":" + itoa(tunnelID)
	c := C
	c.mu.Lock()
	var t TrafficEntry
	if e, ok := c.m[key]; ok {
		if v, ok2 := e.value.(TrafficEntry); ok2 {
			t = v
		}
	}
	t.In += in
	t.Out += out
	c.m[key] = entry{value: t, exp: time.Now().Add(48 * time.Hour)}
	c.mu.Unlock()
}

// GetTraffic 读取隧道当日流量。
func GetTraffic(date string, tunnelID uint) (TrafficEntry, bool) {
	key := "traffic:" + date + ":" + itoa(tunnelID)
	v, ok := C.Get(key)
	if !ok {
		return TrafficEntry{}, false
	}
	t, ok := v.(TrafficEntry)
	return t, ok
}

// DrainTraffic 读取并清零隧道当日流量。
func DrainTraffic(date string, tunnelID uint) (TrafficEntry, bool) {
	key := "traffic:" + date + ":" + itoa(tunnelID)
	v, ok := C.GetAndDel(key)
	if !ok {
		return TrafficEntry{}, false
	}
	t, ok := v.(TrafficEntry)
	return t, ok
}

// TunnelRuntime 隧道实时状态。
type TunnelRuntime struct {
	Online      bool   `json:"online"`
	Connections int    `json:"connections"`
	In          int64  `json:"in"`
	Out         int64  `json:"out"`
	Ts          string `json:"ts"`
}

// RuntimeKey 隧道实时状态 key。
func RuntimeKey(id uint) string { return "tunnel:runtime:" + itoa(id) }

// WantKey 隧道启停指令 key。
func WantKey(id uint) string { return "tunnel:want:" + itoa(id) }

// NodeKey 节点在线标记 key。
func NodeKey(id uint) string { return "node:online:" + itoa(id) }

func itoa(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
