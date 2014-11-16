package fdcache

import (
	"os"
	"sync"

	"github.com/AudriusButkevicius/lfu-go"
)

// A wrapper around *os.File which counts references
type CachedFile struct {
	file *os.File
	wg   sync.WaitGroup
}

// Tells the cache that we are done using the file, but it's up to the cache
// to decide when this file will really be closed. The error, if any, will be
// lost.
func (f *CachedFile) Close() error {
	f.wg.Done()
	return nil
}

// Read the file at the given offset.
func (f *CachedFile) ReadAt(buf []byte, at int64) (int, error) {
	return f.file.ReadAt(buf, at)
}

type FileCache struct {
	cache *lfu.Cache
	mut   sync.RWMutex // Protects against races between concurrent opens
}

// Create a new cache with the given upper and lower LFU limits.
func NewCache(upper, lower int) *FileCache {
	c := FileCache{
		cache: lfu.New(),
	}
	c.cache.UpperBound = upper
	c.cache.LowerBound = lower
	evict := make(chan lfu.Eviction)
	c.cache.EvictionChannel = evict

	go func() {
		for inf := range evict {
			// The file might still be in use, hence spawn a routine to close
			// the file when it has been Close()'d by all openers.
			go func(item *CachedFile) {
				item.wg.Wait()
				item.file.Close()
			}(inf.Value.(*CachedFile))
		}
	}()
	return &c
}

// Open and cache a file descriptor or use an existing cached descriptor for
// the given path.
func (c *FileCache) Open(path string) (*CachedFile, error) {
	// Evictions can only happen during c.cache.Set, and there is a potential
	// race between c.cache.Get and cfd.wg.Add where if not guarded by a mutex
	// could result in cfd getting closed before the counter is incremented if
	// a concurrent routine does a c.cache.Set
	c.mut.RLock()
	fdi := c.cache.Get(path)
	if fdi != nil {
		cfd := fdi.(*CachedFile)
		cfd.wg.Add(1)
		c.mut.RUnlock()
		return cfd, nil
	}
	c.mut.RUnlock()

	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cfd := &CachedFile{
		file: fd,
		wg:   sync.WaitGroup{},
	}
	cfd.wg.Add(1)
	c.mut.Lock()
	c.cache.Set(path, cfd)
	c.mut.Unlock()
	return cfd, nil
}
