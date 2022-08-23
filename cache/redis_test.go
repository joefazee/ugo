package cache

import "testing"

func TestRedisCache_Has(t *testing.T) {

	err := testRedisCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache found but it shouldn`t be there")
	}

	err = testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	inCache, err = testRedisCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Errorf("foo should be in cache %v found", inCache)
	}

}

func TestRedisCache_Get(t *testing.T) {

	err := testRedisCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	x, err := testRedisCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if x != "bar" {
		t.Errorf("did not get correct value from cache; got %v", x)
	}
}

func TestRedisCache_Forget(t *testing.T) {

	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Forget("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha found in cache and shouldn`t be there")
	}

}

func TestRedisCache_Empty(t *testing.T) {

	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Empty()
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha found in cache and shouldn`t be there")
	}
}

func TestRedisCache_EmptyByMatch(t *testing.T) {
	err := testRedisCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("alpha2", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.Set("beta", "alpha")
	if err != nil {
		t.Error(err)
	}

	err = testRedisCache.EmptyByMatch("alpha")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testRedisCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha found in cache and shouldn`t be there")
	}

	inCache, err = testRedisCache.Has("alpha2")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha2 found in cache and shouldn`t be there")
	}

	inCache, err = testRedisCache.Has("beta")
	if err != nil {
		t.Error(err)
	}
	if !inCache {
		t.Error("beta should still be in the cache")
	}
}
