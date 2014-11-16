package lfufdcache

import (
	"io/ioutil"
	"os"
	"sync"

	"testing"
)

func TestNoopReadFailsOnClosed(t *testing.T) {
	fd, err := ioutil.TempFile("", "fdcache")
	if err != nil {
		t.Fatal(err)
		return
	}
	fd.WriteString("test")
	fd.Close()
	buf := make([]byte, 4)
	_, err = fd.ReadAt(buf, 0)
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSingleFileEviction(t *testing.T) {
	c := NewCache(1, 1)

	wg := sync.WaitGroup{}

	fd, err := ioutil.TempFile("", "fdcache")
	if err != nil {
		t.Fatal(err)
		return
	}
	fd.WriteString("test")
	fd.Close()
	buf := make([]byte, 4)

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

			_, err = cfd.ReadAt(buf, 0)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
}

func TestMultifileEviction(t *testing.T) {
	c := NewCache(1, 1)

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
			fd.WriteString("test")
			fd.Close()
			buf := make([]byte, 4)
			defer os.Remove(fd.Name())

			cfd, err := c.Open(fd.Name())
			if err != nil {
				t.Fatal(err)
				return
			}
			defer cfd.Close()

			_, err = cfd.ReadAt(buf, 0)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	wg.Wait()
}

func TestMixedEviction(t *testing.T) {
	c := NewCache(1, 1)

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		go func() {
			fd, err := ioutil.TempFile("", "fdcache")
			if err != nil {
				t.Fatal(err)
				return
			}
			fd.WriteString("test")
			fd.Close()
			buf := make([]byte, 4)

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

					_, err = cfd.ReadAt(buf, 0)
					if err != nil {
						t.Fatal(err)
					}
				}()
			}
		}()
	}

	wg.Wait()
}
