package store

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

type getResult struct {
	value string
	found bool
}

func TestInMemoryStore_Get(t *testing.T) {
	tests := []struct {
		name       string
		setBefore  bool
		ttl        time.Duration
		sleepAfter time.Duration
		key        string
		want       getResult
	}{
		{
			name:      "несуществующий ключ",
			setBefore: false,
			key:       "missing",
			want:      getResult{value: "", found: false},
		},
		{
			name:      "существующий ключ без TTL",
			setBefore: true,
			ttl:       0,
			key:       "forever",
			want:      getResult{value: "hello", found: true},
		},
		{
			name:      "TTL ещё не истёк",
			setBefore: true,
			ttl:       100 * time.Millisecond,
			key:       "soon-alive",
			want:      getResult{value: "hello", found: true},
		},
		{
			name:       "TTL уже истёк",
			setBefore:  true,
			ttl:        10 * time.Millisecond,
			sleepAfter: 20 * time.Millisecond,
			key:        "soon-dead",
			want:       getResult{value: "", found: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryStore()

			if tt.setBefore {
				if err := store.Set(tt.key, "hello", tt.ttl); err != nil {
					t.Fatalf("Set(%s) unexpected error: %v", tt.key, err)
				}
			}
			if tt.sleepAfter > 0 {
				time.Sleep(tt.sleepAfter)
			}

			gotValue, gotFound := store.Get(tt.key)
			if gotValue != tt.want.value || gotFound != tt.want.found {
				t.Errorf("Get(%s) = (%q, %v), want (%q, %v)",
					tt.key, gotValue, gotFound, tt.want.value, tt.want.found)
			}
		})
	}
}

func TestInMemoryStore_Set(t *testing.T) {
	tests := []struct {
		name       string
		setBefore  bool
		ttl        time.Duration
		sleepAfter time.Duration
		key, value string
		want       bool
	}{
		{
			name:  "создание нового не существующего ключа",
			ttl:   1 * time.Second,
			key:   "new_key",
			value: "first_new_key",
			want:  true,
		},
		{
			name:       "создание уже существующего ключа",
			setBefore:  true,
			ttl:        1 * time.Second,
			sleepAfter: 2 * time.Second,
			key:        "new_key_2",
			value:      "second_new_key",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryStore()

			if tt.setBefore {
				store.Set(tt.key, "old_value", 0)
			}

			store.Set(tt.key, tt.value, tt.ttl)

			if tt.sleepAfter > 0 {
				time.Sleep(tt.sleepAfter)
			}

			_, gotFound := store.Get(tt.key)
			if gotFound != tt.want {
				t.Errorf("Set(%s) and after Get(%s) = (%v), want (%v)",
					tt.key, tt.key, gotFound, tt.want)
			}
		})
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setBefore  bool
		getAfter   bool
		sleepAfter time.Duration
		ttl        time.Duration
		key        string
		value      string
		want       bool
	}{
		{
			name: "удаление не существующего ключа",
			key:  "not_existing_key",
			want: false,
		},
		{
			name:      "удаление существующего живого ключа",
			setBefore: true,
			getAfter:  true,
			ttl:       10 * time.Millisecond,
			key:       "existing_alive_key",
			value:     "existing_alive_value",
			want:      true,
		},
		{
			name:       "удаление существующего протухшего ключа",
			setBefore:  true,
			ttl:        10 * time.Millisecond,
			sleepAfter: 20 * time.Millisecond,
			key:        "existing_expired_key",
			value:      "existing_expired_value",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryStore()
			if tt.setBefore {
				store.Set(tt.key, tt.value, tt.ttl)
			}
			if tt.sleepAfter > 0 {
				time.Sleep(tt.sleepAfter)
			}
			got := store.Delete(tt.key)

			if tt.getAfter {
				_, ok := store.Get(tt.key)
				if ok {
					t.Errorf("Get after Delete returned true but want false")
				}
			}

			if got != tt.want {
				t.Errorf("Delete(%s) = (%v), want (%v)",
					tt.key, got, tt.want)
			}
		})
	}
}

func TestInMemoryStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStore()
	var numGoroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				key := "key-" + strconv.Itoa(i)
				store.Set(key, key, 0)
				store.Get(key)
				store.Delete(key)
			}
		}()
	}
	wg.Wait()
}
