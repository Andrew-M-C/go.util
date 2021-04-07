package context

import (
	"context"
	"sync"
	"time"
)

// KVContext implements a context storage with multiple key-value pairs
type KVContext struct {
	parent context.Context
	m      sync.Map
}

// Deadline implements context.Context interface
func (kv *KVContext) Deadline() (time.Time, bool) {
	if kv.parent != nil {
		return kv.parent.Deadline()
	}
	return time.Time{}, false
}

// Done implements context.Context interface
func (kv *KVContext) Done() <-chan struct{} {
	if kv.parent != nil {
		return kv.parent.Done()
	}
	return nil
}

// Err implements context.Context interface
func (kv *KVContext) Err() error {
	if kv.parent != nil {
		return kv.parent.Err()
	}
	return nil
}

// Err implements context.Context interface
func (kv *KVContext) Value(key interface{}) interface{} {
	if v, exist := kv.m.Load(key); exist {
		return v
	}
	if kv.parent != nil {
		return kv.parent.Value(key)
	}
	return nil
}

// WithKVs returns a context.Context implements with multple KV pairs
func WithKVs(kvPairs map[interface{}]interface{}, parent ...context.Context) context.Context {
	kv := &KVContext{}
	if len(parent) > 0 && parent[0] != nil {
		kv.parent = parent[0]
	}

	for k, v := range kvPairs {
		kv.m.Store(k, v)
	}
	return kv
}
