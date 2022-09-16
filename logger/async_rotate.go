package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type async struct {
	lock       *sync.Mutex
	ch         chan string
	oldPath    string // 记录当前的目录
	maxBackups int    // 最多保存多少天的日志
	f          *os.File
	path       string        // 最外层的目录
	size       int           // 记录当前写入的大小
	maxSize    int           // 文件限定的大小
	inform     chan struct{} // 通知要创建新的dir
	l          int           // 记录目录中日志文件的数量
}

// NewAsync ：建议使用这个，性能增加 up up up
func NewAsync(maxBackUp, maxSize int, filePath string) *async {
	r := &async{
		lock:       &sync.Mutex{},
		maxBackups: maxBackUp,
		path:       filePath,
		maxSize:    maxSize * 1024 * 1024,
		inform:     make(chan struct{}),
		ch:         make(chan string),
	}
	go r.informer()
	go delete(r.maxBackups, filePath)
	go r.write()
	return r
}

func (r *async) informer() {
	for {
		next := time.Now().AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
		formatTime, err := time.Parse("2006-01-02 15:04:05", next)
		if err != nil {
			time.Sleep(3600 * time.Second)
		} else {
			now := time.Now().UTC()
			//fmt.Println("now is ", now)
			theTime := time.Duration(formatTime.Unix()-now.Unix()) * time.Second
			//fmt.Println("the time is ", theTime)
			time.Sleep(theTime)
		}
		r.inform <- struct{}{}
	}
}

func (r *async) Write(p []byte) (n int, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.ch <- string(p)
	return len(p), nil
}

func (r *async) write() {
	for v := range r.ch {
		select {
		// 每天创建一个新的目录
		case <-r.inform:
			fmt.Println("inform  >>>>>>>>>>")
			r.createDirWriteLog(v)
			continue
		default:
		}
		if r.f == nil {
			// 第一次运行 查询目录是否存在，如果不存在直接创建即可
			if !r.foundDir() {
				r.createDirWriteLog(v)
				continue
			}
			// 如果目录存在，那么需要找到最后一个文件，判断是否大于限定大小，如果超过创建文件，否则直接写入
			r.foundFile(v)
			continue
		}
		r.size += len(v)
		// 如果小于限定大小，那么就直接写入
		if r.size < r.maxSize {
			r.writeLog(v)
			continue
		}
		// 需要查询文件数量
		r.size = 0
		num := r.foundFileNum()
		if num == 0 {
			r.l = 1
		} else {
			r.l = num + 1
		}
		// 这个是超过限定的大小了，那么就需要写入另外一个文件了
		r.WriteNext(v)
		continue
	}
}

func (r *async) WriteNext(p string) (n int, err error) {
	file := fmt.Sprintf("%v/%v.%v", r.oldPath, time.Now().Format(day), r.l)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	if r.f != nil {
		r.f.Close()
	}
	r.f = f
	return r.writeLog(p)
}

func (r *async) foundDir() bool {
	path := fmt.Sprintf("%v/%v", strings.TrimSuffix(r.path, "/"), time.Now().Format(day))
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func (r *async) foundFileNum() int {
	path := fmt.Sprintf("%v/%v", strings.TrimSuffix(r.path, "/"), time.Now().Format(day))
	fNum, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("read dir err", err)
		return 0
	}
	r.oldPath = path
	return len(fNum)
}

func (r *async) foundFile(p string) (n int, err error) {
	num := r.foundFileNum()
	if num == 0 {
		num = 1
	}
	file := fmt.Sprintf("%v/%v.%v", r.oldPath, time.Now().Format(day), num)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		// 读取错了，就重新创建一个文件
		return r.addNumWriteLog(num, p)
	}
	// 判断如果文件大小超过限定大小，需要创建一个新的文件
	if int(size) > r.maxSize {
		return r.addNumWriteLog(num, p)
	}
	// 文件没有超过限定大小，直接写入
	r.f = f
	r.size = 0
	r.size = int(size) + len(p)
	return r.writeLog(p)
}

func (r *async) addNumWriteLog(num int, p string) (n int, err error) {
	r.l = num + 1
	file1 := fmt.Sprintf("%v/%v.%v", r.oldPath, time.Now().Format(day), r.l)
	f1, err := os.OpenFile(file1, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	if r.f != nil {
		r.f.Close()
	}
	r.f = f1
	r.size = 0
	r.size += len(p)
	return r.writeLog(p)
}

func (r *async) createDirWriteLog(p string) (n int, err error) {
	path := fmt.Sprintf("%v/%v", strings.TrimSuffix(r.path, "/"), time.Now().Format(day))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		fmt.Println("mkdir dir err", err)
		return 0, err
	}
	r.oldPath = path
	file := fmt.Sprintf("%v/%v.%v", r.oldPath, time.Now().Format(day), 1)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	r.l = 1
	if r.f != nil {
		//fmt.Println("close",r.f)
		r.f.Close()
	}
	//fmt.Println("f is",f)
	r.f = f
	r.size = 0
	r.size += len(p)
	return r.writeLog(p)
}

func (r *async) writeLog(p string) (n int, err error) {
	n, err = r.f.WriteString(p)
	if err != nil {
		fmt.Println("write err >>>>>>", err)
		return n, err
	}
	return n, err
}
