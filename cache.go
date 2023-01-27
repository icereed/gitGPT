package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/minio/highwayhash"
)

type Cache struct {
	data    map[uint64]string
	hashKey []byte
	file    string
	lock    sync.RWMutex
}

const HASH_KEY = "C0FFEE030405060708090A0C0FFEEE0FF0E0D0C0B0A090807C0FFEE030201000"

func NewCache(file string) (*Cache, error) {
	c := &Cache{
		data: make(map[uint64]string),
		file: expandPath(file),
	}
	if err := c.Load(); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(HASH_KEY)
	if err != nil {
		fmt.Printf("Cannot decode hex key: %v", err) // add error handling
		return nil, err
	}
	c.hashKey = key
	return c, nil
}

func expandPath(pathToBeExpanded string) string {
	if path.IsAbs(pathToBeExpanded) {
		return pathToBeExpanded
	}

	usr, _ := user.Current()
	dir := usr.HomeDir
	if pathToBeExpanded == "~" {
		// In case of "~", which won't be caught by the "else if"
		pathToBeExpanded = dir
	} else if strings.HasPrefix(pathToBeExpanded, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		pathToBeExpanded = filepath.Join(dir, pathToBeExpanded[2:])
	} else {
		// In case of relative paths
		cwd, _ := os.Getwd()
		pathToBeExpanded = filepath.Join(cwd, pathToBeExpanded)
	}
	return pathToBeExpanded
}

func (c *Cache) Load() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, err := os.Stat(c.file)
	if os.IsNotExist(err) {
		return nil
	}

	b, err := ioutil.ReadFile(c.file)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &c.data); err != nil {
		return err
	}

	return nil
}

func (c *Cache) Write() error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	b, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.file), os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(c.file, b, 0644); err != nil {
		return err
	}

	return nil
}
func (c *Cache) hash(key string) uint64 {
	return highwayhash.Sum64([]byte(key), c.hashKey)
}

func (c *Cache) Get(key string) (string, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, ok := c.data[c.hash(key)]
	return value, ok
}

func (c *Cache) Set(key string, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[c.hash(key)] = value
}
