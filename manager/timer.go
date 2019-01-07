package manager

import "time"

type interval struct {
	*time.Ticker
	duration time.Duration
}

func setInterval(duration time.Duration, callback func()) (t *interval) {
	t = new(interval)
	t.duration = duration
	t.Ticker = time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-t.C:
				callback()
			}
		}
	}()
	return
}
func clearInterval(t *interval) {
	t.Stop()
}

type timeout struct {
	*time.Timer
	time time.Time
}

func setTimeout(duration time.Duration, callback func()) (t *timeout) {
	t = new(timeout)
	t.time = time.Now().Add(duration)
	t.Timer = time.AfterFunc(duration, callback)
	return
}
func clearTimeout(t *timeout) bool {
	t.Stop()
	return true
}
