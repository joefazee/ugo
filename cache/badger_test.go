package cache

import "testing"

func TestBadgerCache_Has(t *testing.T) {
	err := testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo found in cache and it shouldn`t be there!")
	}

	_ = testBadgerCache.Set("foo", "bar")

	inCache, err = testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if !inCache {
		t.Error("foo not found in cache")
	}

	err = testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}
}

func TestBadgerCache_Get(t *testing.T) {
	err := testBadgerCache.Set("foo", "bar")
	if err != nil {
		t.Error(err)
	}

	x, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Error(err)
	}

	if x != "bar" {
		t.Error("did not found the correct value in cache")
	}
}

func TestBadgerCache_Forget(t *testing.T) {

	err := testBadgerCache.Set("foo", "foo")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Forget("foo")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Has("foo")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("foo should have been deleted from the cache!")
	}
}

func TestBadgerCache_Empty(t *testing.T) {

	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Empty()

	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}

	if inCache {
		t.Error("alpha should have been deleted from cache")
	}
}

func TestBadgerCache_EmptyByMatch(t *testing.T) {
	err := testBadgerCache.Set("alpha", "beta")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("alpha2", "beta2")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("bar", "bar")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.EmptyByMatch("a")
	if err != nil {
		t.Error(err)
	}

	inCache, err := testBadgerCache.Has("alpha")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha should have been deleted from cache")
	}

	inCache, err = testBadgerCache.Has("alpha2")
	if err != nil {
		t.Error(err)
	}
	if inCache {
		t.Error("alpha2 should have been deleted from cache")
	}

	inCache, err = testBadgerCache.Has("bar")
	if err != nil {
		t.Error(err)
	}
	if !inCache {
		t.Error("bar should still be in the cache")
	}

}
