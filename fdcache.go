package fdcache

import (
	"os"
	"sync"

	"github.com/AudriusButkevicius/lfu-go"
)

// A locking wrapper around os.File
type CachedFile struct {
	*os.File
	wg sync.WaitGroup
}

// Tells the cache that we are done using the file, but it's up to the cache
// to decide when this file will really be closed.
func (f *CachedFile) Close() error {
	f.wg.Done()
	return nil
}

type FileCache struct {
	cache *lfu.Cache
	mut   sync.Mutex
	wg    sync.WaitGroup
}

func NewFileCache(upper, lower int) *FileCache {
	c := FileCache{
		cache: lfu.New(),
		wg:    sync.WaitGroup{},
	}
	c.cache.UpperBound = upper
	c.cache.LowerBound = lower
	evict := make(chan lfu.Eviction)
	c.cache.EvictionChannel = evict

	go func() {
		for inf := range evict {
			// The file might still be in use, hence spawn a routine to close
			// the file when it has been Close()'d by all openers.
			c.wg.Add(1)
			go func(item *CachedFile) {
				item.wg.Wait()
				item.File.Close()
				c.wg.Done()
			}(inf.Value.(*CachedFile))
		}
	}()
	return &c
}

// Open and cache a file descriptor or use an existing cached descriptor for
// the given path.
func (c *FileCache) Open(path string) (*CachedFile, error) {
	// We can only open one file at a time, in order not to trigger any
	// evictions between c.cache.Get, and cfd.wg.Add
	c.mut.Lock()
	defer c.mut.Unlock()

	fdi := c.cache.Get(path)
	if fdi != nil {
		cfd := fdi.(*CachedFile)
		cfd.wg.Add(1)
		return cfd, nil
	}

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfd := &CachedFile{
		File: fd,
		wg:   sync.WaitGroup{},
	}
	cfd.wg.Add(1)
	c.cache.Set(path, cfd)
	return cfd, nil
}

// Flush the cache, blocks until all opened descriptors are closed by their
// users.
func (c *FileCache) Close() {
	c.mut.Lock()
	c.cache.Evict(c.cache.Len())
	c.mut.Unlock()
	c.wg.Wait()
}
