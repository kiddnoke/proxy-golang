package common

import (
	"net"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestPipe(t *testing.T) {
	var wg sync.WaitGroup
	r, w := Pipe(10)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			str := strconv.Itoa(i)
			w.Write([]byte(str))
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		buf := make([]byte, 10)
		for i := 0; i < 50; i++ {
			n, _ := r.Read(buf)
			t.Logf("%v", string(buf[:n]))
		}
		time.Sleep(time.Second)
		for i := 0; i < 50; i++ {
			n, _ := r.Read(buf)
			t.Logf("%v", string(buf[:n]))
		}
	}()
	wg.Wait()
}
func TestIoPipe(t *testing.T) {
	var wg sync.WaitGroup
	r, w := net.Pipe()
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			str := strconv.Itoa(i)
			w.Write([]byte(str))
		}
	}()
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		buf := make([]byte, 10)
		for i := 0; i < 50; i++ {
			n, _ := r.Read(buf)
			t.Logf("%v", string(buf[:n]))
		}
		time.Sleep(time.Second)
		for i := 0; i < 50; i++ {
			n, _ := r.Read(buf)
			t.Logf("%v", string(buf[:n]))
		}
	}()
	wg.Wait()
}
func TestPipe_Block(t *testing.T) {
	r, w := Pipe(10)

	w.Write([]byte("1"))
	w.Write([]byte("2"))
	w.Write([]byte("3"))
	buf := make([]byte, 20)
	r.Read(buf)
	r.Read(buf)
	r.Read(buf)
	r.Read(buf)
}
func TestPipe_Chan(t *testing.T) {
	c := make(chan bool, 20)
	t.Logf("%v", len(c))
	t.Logf("%v", cap(c))
}
func TestpipeDeadline(t *testing.T) {

}
