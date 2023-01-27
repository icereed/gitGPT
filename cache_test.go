package main

// unit tests:
// 1. TestNewCache
// 2. TestLoad
// 3. TestWrite
// 4. TestHash
// 5. TestGet
// 6. TestSet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCache(t *testing.T) {
	// arrange
	cacheFile := filepath.Join(os.TempDir(), "cache.json")

	// act
	cache, err := NewCache(cacheFile)

	// assert
	assert.NoError(t, err)
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.data)
	assert.NotNil(t, cache.hashKey)
	assert.NotNil(t, cache.file)
}

func TestLoad(t *testing.T) {
	// arrange
	cacheFile := filepath.Join(os.TempDir(), "cache.json")
	cache, _ := NewCache(cacheFile)

	// act
	err := cache.Load()

	// assert
	assert.NoError(t, err)
}

func TestWrite(t *testing.T) {
	// arrange
	cacheFile := filepath.Join(os.TempDir(), "cache.json")
	cache, _ := NewCache(cacheFile)
	cache.data[123456] = "test"

	// act
	err := cache.Write()

	// assert
	assert.NoError(t, err)

	// clean up
	os.Remove(cacheFile)
}

func TestSetGet(t *testing.T) {
	// arrange
	cacheFile := filepath.Join(os.TempDir(), "cache.json")
	cache, _ := NewCache(cacheFile)

	// act
	cache.Set("key", "test")

	// assert
	result, ok := cache.Get("key")
	assert.Equal(t, "test", result)
	assert.True(t, ok)
}
