package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pickIQ/oredis/internal/store"
)

type getResult struct {
	value  valueResponse
	status int
}

func TestHttpHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		setBefore  bool
		sleepAfter time.Duration
		key        string
		value      string
		ttl        time.Duration
		want       getResult
	}{
		{
			name:      "ключ существует",
			setBefore: true,
			key:       "existingKey",
			value:     "bar",
			want:      getResult{status: http.StatusOK, value: valueResponse{Value: "bar"}},
		},
		{
			name: "ключ не существует",
			key:  "notFound",
			want: getResult{status: http.StatusNotFound},
		},
		{
			name:       "ключ существует, но ttl истек",
			setBefore:  true,
			ttl:        3 * time.Millisecond,
			sleepAfter: 4 * time.Millisecond,
			key:        "expiredTTL",
			value:      "expiredTTLVal",
			want:       getResult{status: http.StatusNotFound},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := store.NewInMemoryStore()

			router := NewRouter(st)

			if tt.setBefore {
				if err := st.Set(tt.key, tt.value, tt.ttl); err != nil {
					t.Fatalf("Set(%s) unexpected error: %v", tt.key, err)
				}
			}

			if tt.sleepAfter > 0 {
				time.Sleep(tt.sleepAfter)
			}

			req := httptest.NewRequest(http.MethodGet, "/keys/"+tt.key, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.want.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.want.status)
			}

			if tt.want.status == http.StatusOK {
				var got valueResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("failed to parse response body: %v", err)
				}

				if got != tt.want.value {
					t.Errorf("body = %+v, want %+v", got, tt.want.value)
				}
			}
		})
	}
}

func TestHttpHandler_Put(t *testing.T) {

	tests := []struct {
		name        string
		getAfter    bool
		invalidBody bool
		key         string
		value       string
		ttl         int
		want        int
	}{
		{
			name:     "валидный JSON c value",
			getAfter: true,
			key:      "existingKey",
			value:    "existingValue",
			ttl:      0,
			want:     http.StatusNoContent,
		},
		{
			name:  "валидный JSON c TTL",
			key:   "existingKey2",
			value: "existingValue2",
			ttl:   30,
			want:  http.StatusNoContent,
		},
		{
			name:        "невалидный JSON",
			invalidBody: true,
			key:         "notExistingKey",
			value:       "notExistingValue",
			ttl:         45,
			want:        http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := store.NewInMemoryStore()
			router := NewRouter(st)
			var body io.Reader

			if tt.invalidBody {
				body = strings.NewReader(`{"value":"hellttl_seconds":30}`)
			} else {
				body = strings.NewReader(fmt.Sprintf(`{"value":%q,"ttl_seconds":%d}`, tt.value, tt.ttl))
			}

			req := httptest.NewRequest(http.MethodPut, "/keys/"+tt.key, body)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}

			if tt.getAfter {
				if _, found := st.Get(tt.key); !found {
					t.Errorf("after getAfter, want %v,  found = %v ", tt.value, found)
				}
			}
		})
	}

}

func TestHttpHandler_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setBefore bool
		getAfter  bool
		key       string
		want      int
	}{
		{
			name:      "ключ существует",
			setBefore: true,
			getAfter:  true,
			key:       "existingKey",
			want:      http.StatusNoContent,
		},
		{
			name: "ключ не существует",
			key:  "notExistingKey",
			want: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := store.NewInMemoryStore()
			router := NewRouter(st)

			if tt.setBefore {
				if err := st.Set(tt.key, "some value", 0); err != nil {
					t.Fatalf("Set(%s) unexpected error: %v", tt.key, err)
				}
			}
			req := httptest.NewRequest(http.MethodDelete, "/keys/"+tt.key, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d", rec.Code, tt.want)
			}

			if tt.getAfter {
				if _, found := st.Get(tt.key); found {
					t.Errorf("after getAfter, want %v,  found = %v ", false, found)
				}
			}

		})
	}
}
