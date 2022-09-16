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

var (
	day = "2006-01-02"
)

type Rotate struct {
	oldPath    string // 记录当前的目录
	maxBackups int    // 最多保存多少天的日志
	f          *os.File
	path       string // 最外层的目录
	mu         *sync.Mutex
	size       int           // 记录当前写入的大小
	maxSize    int           // 文件限定的大小
	inform     chan struct{} // 通知要创建新的dir
	l          int           // 记录目录中日志文件的数量
}

func NewRotate(maxBackUp, maxSize int, filePath string) *Rotate {
	r := &Rotate{
		maxBackups: maxBackUp,
		path:       filePath,
		mu:         &sync.Mutex{},
		maxSize:    maxSize * 1024 * 1024,
		inform:     make(chan struct{}),
	}
	go r.informer()
	go delete(r.maxBackups, filePath)
	return r
}

func (r *Rotate) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	select {
	// 每天创建一个新的目录
	case <-r.inform:
		fmt.Println("inform  >>>>>>>>>>")
		return r.createDirWriteLog(p)
	default:
	}

	if r.f == nil {
		// 第一次运行 查询目录是否存在，如果不存在直接创建即可
		if !r.foundDir() {
			return r.createDirWriteLog(p)
		}
		// 如果目录存在，那么需要找到最后一个文件，判断是否大于限定大小，如果超过创建文件，否则直接写入
		return r.foundFile(p)
	}
	r.size += len(p)
	// 如果小于限定大小，那么就直接写入
	if r.size < r.maxSize {
		return r.writeLog(p)
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
	return r.WriteNext(p)
}

func (r *Rotate) WriteNext(p []byte) (n int, err error) {
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

func (r *Rotate) foundDir() bool {
	path := fmt.Sprintf("%v/%v", strings.TrimSuffix(r.path, "/"), time.Now().Format(day))
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func (r *Rotate) foundFileNum() int {
	path := fmt.Sprintf("%v/%v", strings.TrimSuffix(r.path, "/"), time.Now().Format(day))
	fNum, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println("read dir err", err)
		return 0
	}
	r.oldPath = path
	return len(fNum)
}

func (r *Rotate) foundFile(p []byte) (n int, err error) {
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

func (r *Rotate) addNumWriteLog(num int, p []byte) (n int, err error) {
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

func (r *Rotate) createDirWriteLog(p []byte) (n int, err error) {
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
		fmt.Println("close", r.f)
		r.f.Close()
	}
	fmt.Println("f is", f)
	r.f = f
	r.size = 0
	r.size += len(p)
	return r.writeLog(p)
}

func (r *Rotate) writeLog(p []byte) (n int, err error) {
	//now := time.Now()
	n, err = r.f.Write(p)
	if err != nil {
		fmt.Println("write err >>>>>>", err)
		return n, err
	}
	//fmt.Println("---",time.Since(now))
	return n, err
}

func (r *Rotate) informer() {
	for {
		next := time.Now().AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
		fmt.Println("next is >>>", next)
		formatTime, err := time.Parse("2006-01-02 15:04:05", next)
		fmt.Println("format time is ", formatTime)
		if err != nil {
			time.Sleep(3600 * time.Second)
		} else {
			now := time.Now().UTC()
			//fmt.Println("now is ", now)
			theTime := time.Duration(formatTime.Unix()-now.Unix()) * time.Second
			//fmt.Println("the time is ", theTime)
			time.Sleep(theTime)
		}
		//time.Sleep(10*time.Second)
		r.inform <- struct{}{}
	}
}

// 定时删除目录
func delete(maxBackups int, path string) {
	defer fmt.Println("[delete] end")
	fmt.Println("maxBackups", maxBackups, "path", path)
	if maxBackups <= 0 {
		return
	}
	for {
		now := time.Now()
		before := now.AddDate(0, 0, -maxBackups)
		b := before.Format("2006-01-02")
		next := time.Now().AddDate(0, 0, 1).Format("2006-01-02 00:00:00")
		formatTime, err := time.ParseInLocation("2006-01-02 15:04:05", next, time.Local)
		if err != nil {
			fmt.Println("err>>>", err)
			time.Sleep(12 * time.Hour)
		}
		time.Sleep(time.Duration(formatTime.Unix()-time.Now().Unix()) * time.Second)
		findAndDelDir(b, path)
	}
}

func findAndDelDir(beforeFile, path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Println("mkdir  dir err", err)
			return
		}
	}
	for k := range files {
		// 校验是否为时间字符串
		if _, err := time.Parse(day, files[k].Name()); err != nil {
			continue
		}
		if files[k].Name() <= beforeFile {
			if err := os.RemoveAll(fmt.Sprintf("%v/%v", strings.TrimSuffix(path, "/"), files[k].Name())); err != nil {
				fmt.Println("[findAndDelDir] remove file err", err)
			}
		}
	}
}
