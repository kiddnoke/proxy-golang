package manager

import "time"

type interval struct {
	*time.Ticker
}

func setInterval(duration time.Duration, callback func(when time.Time)) (t *interval) {
	t = new(interval)
	t.Ticker = time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-t.C:
				callback(time.Now())
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
}

func setTimeout(duration time.Duration, callback func(when time.Time)) (t *timeout) {
	t = new(timeout)
	t.Timer = time.AfterFunc(duration, func() {
		callback(time.Now())
	})
	return
}
func clearTimeout(t *timeout) bool {
	return t.Stop()
}
