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

	defer func() {
		err := os.Remove(fd.Name())
		if err != nil {
			t.Fatal(err)
		}
	}()

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

			_, err = cfd.Write([]byte(cfd.Name()))
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
	c.Close()
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

			_, err = cfd.Write([]byte(cfd.Name()))
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
	c.Close()
}

func TestMixedEviction(t *testing.T) {
	c := NewFileCache(1, 1)

	wg := sync.WaitGroup{}
	for i := 0; i < 30; i++ {
		fd, err := ioutil.TempFile("", "fdcache")
		if err != nil {
			t.Fatal(err)
			return
		}
		fd.Close()

		defer func() {
			err := os.Remove(fd.Name())
			if err != nil {
				t.Fatal(err)
			}
		}()

		for k := 0; k < 50; k++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				cfd, err := c.Open(fd.Name())
				if err != nil {
					t.Fatal(err)
					return
				}
				defer cfd.Close()

				_, err = cfd.Write([]byte(cfd.Name()))
				if err != nil {
					t.Fatal(err)
				}
			}()
		}
	}

	wg.Wait()
	c.Close()
}
