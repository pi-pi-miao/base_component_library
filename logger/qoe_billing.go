package logger

import (
	"fmt"
	"io/fs"
	"os"
	"sync"
	"time"
)

type QoeBilling struct {
	maxBackups int
	oldPath    string
	file       *os.File
	path       string
	informCh   chan struct{}
	mu         *sync.Mutex
}

func NewQoeBilling(path, name string, maxBackups int) *QoeBilling {
	r := &QoeBilling{
		maxBackups: maxBackups,
		path:       path,
		informCh:   make(chan struct{}),
		mu:         &sync.Mutex{},
	}
	if _, err := os.Stat(r.path); err != nil {
		os.MkdirAll(r.path, fs.ModePerm)
	}
	r.oldPath = fmt.Sprintf("%v/%v.log", path, name)
	f, err := os.OpenFile(r.oldPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("open file err", err)
	}
	r.file = f
	go r.timer()
	return r
}

func (r *QoeBilling) timer() {
	h, err := time.ParseDuration("1h")
	if err != nil {
		fmt.Println("parse duration err")
		return
	}
	for {
		now := time.Now().Format("2006-01-02 15:00:00")
		t, err := time.ParseInLocation("2006-01-02 15:04:05", now, time.Local)
		if err != nil {
			fmt.Println("parse local err", err)
			Errorf("parse local time err:[%v] >>>>>>>", err)
			time.Sleep(1 * time.Hour)
			continue
		}
		t1 := t.Add(h).Unix() - time.Now().Unix()
		time.Sleep(time.Duration(t1) * time.Second)
		r.informCh <- struct{}{}
	}
}

func (r *QoeBilling) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	case <-r.informCh:
		newPath := fmt.Sprintf("%v-%v", r.oldPath, time.Now().Format("2006-010215"))
		if err := os.Rename(r.oldPath, newPath); err != nil {
			Errorf("[rename] err:[%v] >>>>>>>>>>", err)
		} else {
			f, err := os.OpenFile(r.oldPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				Errorf("[utils-logger-writeFile] err:[%v]", err)
				return len(p), err
			}
			if r.file != nil {
				r.file.Close()
			}
			r.file = f
		}
	default:
	}
	if _, err := r.file.Write(p); err != nil {
		fmt.Println("[utils-logger-writeFile] err", err, "data is", string(p))
	}
	return len(p), nil
}

func (r *QoeBilling) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}
