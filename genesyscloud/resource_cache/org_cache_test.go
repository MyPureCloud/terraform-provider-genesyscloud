package resource_cache

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mypurecloud/terraform-provider-genesyscloud/genesyscloud/tfexporter_state"
)

func TestUnitOrgCacheInactiveExporter(t *testing.T) {
	tfexporter_state.ResetExporterStateForTests()
	t.Cleanup(tfexporter_state.ResetExporterStateForTests)

	cache := NewOrgCache(OrgCacheConfig[int]{
		Name:    "test-cache",
		KeyFunc: func(v int) string { return "key" },
		LoadFunc: func(ctx context.Context) ([]int, error) {
			t.Fatal("load should not be called when exporter is inactive")
			return nil, nil
		},
	})

	if err := cache.EnsureLoaded(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if item, ok := cache.Get("key"); ok || item != nil {
		t.Fatalf("expected cache miss when exporter inactive")
	}

	if cache.IsLoaded() {
		t.Fatal("expected cache to report not loaded when exporter inactive")
	}
}

func TestUnitOrgCacheLoadOnce(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	var loadCount atomic.Int32
	cache := NewOrgCache(OrgCacheConfig[int]{
		Name: "test-cache",
		KeyFunc: func(v int) string {
			return "key"
		},
		LoadFunc: func(ctx context.Context) ([]int, error) {
			loadCount.Add(1)
			return []int{42}, nil
		},
	})

	ctx := context.Background()
	if err := cache.EnsureLoaded(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := cache.EnsureLoaded(ctx); err != nil {
		t.Fatalf("unexpected error on second load: %v", err)
	}

	if loadCount.Load() != 1 {
		t.Fatalf("expected load func to run once, got %d", loadCount.Load())
	}

	item, ok := cache.Get("key")
	if !ok || item == nil || *item != 42 {
		t.Fatalf("expected cached value 42, got %v ok=%v", item, ok)
	}

	hits, misses := cache.Stats()
	if hits != 1 || misses != 0 {
		t.Fatalf("expected 1 hit and 0 misses, got hits=%d misses=%d", hits, misses)
	}

	if _, ok := cache.Get("missing"); ok {
		t.Fatal("expected miss for missing key")
	}
	hits, misses = cache.Stats()
	if hits != 1 || misses != 1 {
		t.Fatalf("expected 1 hit and 1 miss, got hits=%d misses=%d", hits, misses)
	}
}

func TestUnitOrgCacheConcurrentLoad(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	var loadCount atomic.Int32
	startGate := make(chan struct{})

	cache := NewOrgCache(OrgCacheConfig[string]{
		Name: "test-cache",
		KeyFunc: func(v string) string {
			return v
		},
		LoadFunc: func(ctx context.Context) ([]string, error) {
			loadCount.Add(1)
			<-startGate
			return []string{"a", "b"}, nil
		},
	})

	ctx := context.Background()
	const workers = 8
	var wg sync.WaitGroup
	wg.Add(workers)

	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			errs <- cache.EnsureLoaded(ctx)
		}()
	}

	close(startGate)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if loadCount.Load() != 1 {
		t.Fatalf("expected load func to run once under concurrency, got %d", loadCount.Load())
	}

	if cache.Size() != 2 {
		t.Fatalf("expected cache size 2, got %d", cache.Size())
	}
}

func TestUnitOrgCacheEnsureLoadedOverride(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	cache := NewOrgCache(OrgCacheConfig[int]{
		Name: "test-cache",
		KeyFunc: func(v int) string {
			return "key"
		},
	})

	err := cache.EnsureLoaded(context.Background(), func(ctx context.Context) ([]int, error) {
		return []int{99}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	item, ok := cache.Get("key")
	if !ok || item == nil || *item != 99 {
		t.Fatalf("expected cached override value 99, got %v ok=%v", item, ok)
	}
}

func TestUnitOrgCacheLoadError(t *testing.T) {
	tfexporter_state.ActivateExporterState()

	cache := NewOrgCache(OrgCacheConfig[int]{
		Name: "test-cache",
		KeyFunc: func(v int) string {
			return "key"
		},
		LoadFunc: func(ctx context.Context) ([]int, error) {
			return nil, errors.New("load failed")
		},
	})

	err := cache.EnsureLoaded(context.Background())
	if err == nil {
		t.Fatal("expected load error")
	}

	if cache.IsLoaded() {
		t.Fatal("expected cache to remain unloaded after error")
	}
}
