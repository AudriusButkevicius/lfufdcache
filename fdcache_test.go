package fdcache

import (
	"io/ioutil"
	"os"
	"sync"

	"testing"
)

func TestSingleFileEviction(t *testing.T) {
	c := NewFileCache(1, 1)

	wg := sync.WaitGroup{}

	fd, err := ioutil.TempFile("", "fdcache")
	if err != nil {
		t.Fatal(err)
		return
	}
	fd.Close()

	for k := 0; k < 100; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			cfd, err := c.Open(fd.Name())
			if err != nil {
				t.Fatal(err)
				return
			}
			defer cfd.Close()

			_, err = cfd.ReadAt([]byte{}, 0)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
}

func TestMultifileEviction(t *testing.T) {
	c := NewFileCache(1, 1)

	wg := sync.WaitGroup{}

	for k := 0; k < 100; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			fd, err := ioutil.TempFile("", "fdcache")
			if err != nil {
				t.Fatal(err)
				return
			}
			fd.Close()
			defer os.Remove(fd.Name())

			cfd, err := c.Open(fd.Name())
			if err != nil {
				t.Fatal(err)
				return
			}
			defer cfd.Close()

			_, err = cfd.ReadAt([]byte{}, 0)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
}

func TestMixedEviction(t *testing.T) {
	c := NewFileCache(1, 1)

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		fd, err := ioutil.TempFile("", "fdcache")
		if err != nil {
			t.Fatal(err)
			return
		}
		fd.Close()

		for k := 0; k < 100; k++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				cfd, err := c.Open(fd.Name())
				if err != nil {
					t.Fatal(err)
					return
				}
				defer cfd.Close()

				_, err = cfd.ReadAt([]byte{}, 0)
				if err != nil {
					t.Fatal(err)
				}
			}()
		}
	}

	wg.Wait()
}
