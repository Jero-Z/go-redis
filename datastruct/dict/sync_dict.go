package dict

import "sync"

type SyncDict struct {
	m sync.Map
}

func (s *SyncDict) Get(key string) (val interface{}, exists bool) {
	val, exists = s.m.Load(key)
	return
}

func (s *SyncDict) Len() int {
	var length int
	s.m.Range(func(key, value any) bool {
		length++
		return true
	})
	return length
}

func (s *SyncDict) Put(key string, val interface{}) (result int) {
	_, ok := s.m.Load(key)
	s.m.Store(key, val)
	if ok {
		return 0
	}
	return 1
}

func (s *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {

	if _, ok := s.m.Load(key); !ok {
		return 0
	}
	s.m.Store(key, val)
	return 1
}

func (s *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	if _, ok := s.m.Load(key); ok {
		s.m.Store(key, val)
		return 1
	}
	return 0
}

func (s *SyncDict) Remove(key string) (result int) {
	_, ok := s.m.Load(key)
	if !ok {
		return 0
	}
	s.m.Delete(key)
	return 1
}

func (s *SyncDict) Foreach(consumer Consumer) {
	s.m.Range(func(key, value any) bool {
		consumer(key.(string), value)
		return true
	})
}

func (s *SyncDict) Keys() []string {
	var result = make([]string, s.Len())
	var i int
	s.m.Range(func(key, value any) bool {
		result[i] = key.(string)
		i++
		return true
	})
	return result
}

func (s *SyncDict) RandomKeys(limit int) []string {
	if s.Len() < limit {

		// db 中存储的个数小于当前需要的数量
		var result = make([]string, s.Len())
		i := 0
		s.m.Range(func(key, value any) bool {
			result[i] = key.(string)
			i++
			return true
		})
		return result
	}

	var result = make([]string, limit)
	for i := 0; i < limit; i++ {
		s.m.Range(func(key, value any) bool {
			result[i] = key.(string)
			return false
		})
	}
	return result
}

func (s *SyncDict) RandomDistinctKeys(limit int) []string {
	// 先查看db 内部存储的数据个数是否大于客户端要的数量，如果小于则将客户端需要的数量修改为当前map中元素的个数
	if s.Len() < limit {
		limit = s.Len()
	}
	// 通过修改后的需求数量建立一个指定长度的slice
	var result = make([]string, limit)
	i := 0
	s.m.Range(func(key, value any) bool {
		result[i] = key.(string)
		i++
		if i == limit {
			return false
		}
		return true
	})
	return result
}

func (s *SyncDict) Clear() {
	*s = *NewSyncDict()
}

func NewSyncDict() *SyncDict {
	return &SyncDict{}
}
