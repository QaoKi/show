package log

// Basic examples:
//
//	xlog.Inf("Prepare to repel boarders")
//
//	xlog.Ftl("Initialization failed: %s", err)
//
// Log output is buffered and written periodically using Flush. Programs
// should call Flush before exiting to guarantee all log output is written.
//
// By default, all log statements write to files in a temporary directory.
// This package provides several flags that modify this behavior.
// As a result, flag.Parse must be called before any logging is done.

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	digits = "0123456789"
	// bufferSize sizes the buffer associated with each log file. It's large
	// so that log records can accumulate without the logging thread blocking
	// on disk I/O. The flushDaemon will block instead.
	bufferSize = 1 << 16
)

const (
	debugLog uint = iota
	infoLog
	warningLog
	errorLog
	clientLog // 远程日志单独存文件
	fatalLog
	dbLog
	maxLogCount // 用于日志等级种类计数
)

// 日志配置
type LogConfig struct {
	SepDbgLog     bool   // 是否分离debug日志到.dbg文件
	LogDir        string // 日志目录，default: ./log
	LogName       string // 日志名，default: 启动程序名
	Hourly        bool   // 按小时滚动，default: daily
	FuncCallDepth int    // 打印调用文件，default: 0
	MaskMap       bool   // 屏蔽打印map内容，防止多线程宕机，default: false
	GzipFileDays  int    // 压缩老旧文件, 默认0，不压缩
	DelFileDays   int    // 删除老旧文件, 默认0，不删除
	Console       bool   // 是否打印到控制台

	rotateInterval int
}

func (c *LogConfig) init() {
	c.LogDir = "./log"
	c.LogName = filepath.Base(os.Args[0])
}

type logInfo struct {
	logType uint   //日志类型 log or dbg?
	postfix string //日志后缀
}

var (
	zoneOffset int

	severityChar = []string{"DBG", "INF", "WRN", "ERR", "CLI", "FTL"}
	logInfoList  []logInfo
	closeLog     bool
	waitForMq    int64

	conf    LogConfig
	logging loggingT
)

func Init(jsonConfig string) {
	_, offset := time.Now().Zone()
	zoneOffset = offset

	conf.init()
	if len(jsonConfig) == 0 {
		jsonConfig = "{}"
	}
	err := json.Unmarshal([]byte(jsonConfig), &conf)
	if err != nil {
		fmt.Printf("hllog config error:%v\n", err)
	}

	//初始化入口
	//日志等级     => 日志类型 + 后缀
	//debug        => 1
	//info...fatal => 0
	logInfoList = make([]logInfo, 0, maxLogCount)
	for i := uint(0); i < maxLogCount; i++ {
		logInfoList = append(logInfoList, logInfo{logType: 0, postfix: ".log"})
	}
	if conf.SepDbgLog {
		logInfoList[debugLog] = logInfo{logType: 1, postfix: ".dbg"}
	}
	logInfoList[clientLog] = logInfo{logType: 1, postfix: ".clg"}
	conf.rotateInterval = 3600
	if !conf.Hourly {
		conf.rotateInterval *= 24
	}

	logging.rawMq = make(chan rawInput, 1<<16)

	go logging.flushDaemon()
	go logging.format()
	go logging.idleTimer()
}

func Flush() {
	flush()
}

func flush() { logging.lockAndFlushAll() }

func Final() {
	var raw rawInput
	raw.s = maxLogCount
	logging.rawMq <- raw
	for atomic.LoadInt64(&waitForMq) == 0 {
		time.Sleep(1e7)
	}
	flush()
}

func Config() LogConfig { return conf }

func SetClose(b bool) { closeLog = b }

func RedirectStderr() {
	errFileName := "/err." + conf.LogName
	errFile, _ := os.OpenFile(conf.LogDir+errFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE|os.O_SYNC, 0644)
	syscall.Dup2(int(errFile.Fd()), 2)
}

// loggingT collects all the global state of the logging setup.
type loggingT struct {
	freeList   *buffer
	freeListMu sync.Mutex

	mu   sync.Mutex
	file [maxLogCount]*syncBuffer

	rawMq chan rawInput
}

type rawInput struct {
	s      uint
	now    time.Time
	format string
	caller string
	args   []interface{}
}

// buffer holds a byte Buffer for reuse. The zero value is ready for use.
type buffer struct {
	bytes.Buffer
	tmp  [64]byte // temporary byte array for creating headers.
	next *buffer
}

// getBuffer returns a new, ready-to-use buffer.
func (l *loggingT) getBuffer() *buffer {
	l.freeListMu.Lock()
	b := l.freeList
	if b != nil {
		l.freeList = b.next
	}
	l.freeListMu.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the free list.
func (l *loggingT) putBuffer(b *buffer) {
	if b.Len() >= 256 {
		// Let big buffers die a natural death.
		return
	}
	l.freeListMu.Lock()
	b.next = l.freeList
	l.freeList = b
	l.freeListMu.Unlock()
}

// formatHeader formats a log header using the provided file name and line number.
func (l *loggingT) header(now time.Time, s uint) *buffer {
	if s > fatalLog {
		s = debugLog
	}
	buf := l.getBuffer()

	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	// Lmmdd hh:mm:ss.uuuuuu]
	buf.nDigits(4, 0, year, '0')
	buf.twoDigits(4, int(month))
	buf.twoDigits(6, day)
	buf.tmp[8] = ' '
	buf.twoDigits(9, hour)
	buf.tmp[11] = ':'
	buf.twoDigits(12, minute)
	buf.tmp[14] = ':'
	buf.twoDigits(15, second)
	buf.tmp[17] = '.'
	buf.nDigits(4, 18, now.Nanosecond()/1e5, '0')
	buf.tmp[22] = ' '
	buf.Write(buf.tmp[:23])

	str := severityChar[s]
	buf.tmp[0] = str[0]
	buf.tmp[1] = str[1]
	buf.tmp[2] = str[2]
	buf.tmp[3] = ':'
	buf.Write(buf.tmp[:4])
	return buf
}

func (buf *buffer) twoDigits(i, d int) {
	buf.tmp[i+1] = digits[d%10]
	d /= 10
	buf.tmp[i] = digits[d%10]
}

// nDigits formats an n-digit integer at buf.tmp[i],
// padding with pad on the left.
// It assumes d >= 0.
func (buf *buffer) nDigits(n, i, d int, pad byte) {
	j := n - 1
	for ; j >= 0 && d > 0; j-- {
		buf.tmp[i+j] = digits[d%10]
		d /= 10
	}
	for ; j >= 0; j-- {
		buf.tmp[i+j] = pad
	}
}

// someDigits formats a zero-prefixed variable-width integer at buf.tmp[i].
func (buf *buffer) SomeDigits(i, d int) int {
	// Print into the top, then copy down. We know there's space for at least
	// a 10-digit number.
	j := len(buf.tmp)
	for {
		j--
		buf.tmp[j] = digits[d%10]
		d /= 10
		if d == 0 {
			break
		}
	}
	return copy(buf.tmp[i:], buf.tmp[j:])
}

func (l *loggingT) printf(s uint, format string, args ...interface{}) {
	if conf.Console {
		fmt.Printf("["+severityChar[s]+"] "+format+"\n", args...)
	}
	if closeLog {
		return
	}
	var raw rawInput
	raw.s = s
	raw.now = time.Now()
	raw.format = format
	raw.args = args
	if conf.FuncCallDepth > 0 {
		if _, file, line, ok := runtime.Caller(conf.FuncCallDepth); ok {
			_, filename := path.Split(file)
			raw.caller = fmt.Sprintf("[%v:%v],", filename, line)
		}
	}

	select {
	case l.rawMq <- raw:
	default:
		fmt.Println(s, format, args)
	}
}

func (l *loggingT) format() {
	for {
		raw := <-l.rawMq
		if raw.s != maxLogCount {
			buf := l.header(raw.now, raw.s)
			if len(raw.caller) > 0 {
				buf.WriteString(raw.caller)
			}
			if conf.MaskMap {
				for i, v := range raw.args {
					if reflect.ValueOf(v).Kind() == reflect.Map {
						raw.args[i] = "maskedMap"
					}
				}
			}
			fmt.Fprintf(buf, raw.format, raw.args...)
			buf.WriteByte('\n')
			l.output(raw.s, buf)
		} else {
			atomic.StoreInt64(&waitForMq, -1)
			return
		}
	}
}

// 根据日志对应等级==>文件类型 "0/1"+ 文件后缀, ".log/.dbg"
func (l *loggingT) getFileDespInfo(s uint) (uint, string) {
	return logInfoList[s].logType, logInfoList[s].postfix
}

// output writes the data to the log files and releases the buffer.
func (l *loggingT) output(s uint, buf *buffer) {
	l.mu.Lock()
	data := buf.Bytes()
	idx, _ := l.getFileDespInfo(s)
	if l.file[idx] == nil {
		if err := l.createFile(s); err != nil {
			os.Stderr.Write(data) // Make sure the message appears somewhere.
			l.exit(err)
		}
	}
	l.file[idx].Write(data)

	if s == fatalLog {
		// Write the stack trace for all goroutines to the files.
		trace := stacks(true)
		logExitFunc = func(error) {}    // If we get a write error, we'll still exit below.
		if f := l.file[idx]; f != nil { // Can be nil if -logtostderr is set.
			f.Write(trace)
		}
		l.mu.Unlock()
		flush()
		os.Exit(255) // C++ uses -1, which is silly because it's anded with 255 anyway.
	}

	l.putBuffer(buf)
	l.mu.Unlock()
}

func stacks(all bool) []byte {
	n := 1 << 10
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		num := runtime.Stack(trace, all)
		if num < len(trace) {
			return trace[:num]
		}
		n *= 8
	}
	return trace
}

var logExitFunc func(error)

// exit is called if there is trouble creating or writing log files.
// It flushes the logs and exits the program; there's no point in hanging around.
// l.mu is held.
func (l *loggingT) exit(err error) {
	fmt.Fprintf(os.Stderr, "log: exiting because of error: %s\n", err)
	if logExitFunc != nil {
		logExitFunc(err)
		return
	}
	l.flushAll()
	os.Exit(2)
}

// syncBuffer joins a bufio.Writer to its underlying file, providing access to the
// file's Sync method and providing a wrapper for the Write method that provides log
// file rotation. There are conflicting methods, so the file cannot be embedded.
// l.mu is held for all its methods.
type syncBuffer struct {
	logger *loggingT
	*bufio.Writer
	osfile *os.File
	nbytes uint64 // The number of bytes written to this file
	nlines int    // new lines
	sbName string

	lastRotateTick int
}

func (sb *syncBuffer) Sync() error { return sb.osfile.Sync() }

func (sb *syncBuffer) Write(p []byte) (n int, err error) {
	now := time.Now()
	curTick := rotateTick(now)
	if curTick != sb.lastRotateTick {
		if err := sb.rotateFile(now); err != nil {
			sb.logger.exit(err)
		}
		sb.lastRotateTick = curTick
	}
	n, err = sb.Writer.Write(p)
	sb.nbytes += uint64(n)
	sb.nlines++
	if err != nil {
		sb.logger.exit(err)
	}
	return
}

func (sb *syncBuffer) rotateFile(now time.Time) error {
	if sb.osfile != nil {
		sb.Flush()
		sb.osfile.Close()
		sb.nlines = 0
	}
	var err error
	sb.osfile, _, err = setFile(now, sb.sbName)
	sb.nbytes = 0
	if err != nil {
		return err
	}

	sb.Writer = bufio.NewWriterSize(sb.osfile, bufferSize)

	var buf bytes.Buffer
	n, err := sb.osfile.Write(buf.Bytes())
	sb.nbytes += uint64(n)
	sb.nlines++
	return err
}

func logName(t time.Time, prefix string) (name string) {
	name = fmt.Sprintf("%s.%04d%02d%02d", prefix, t.Year(), t.Month(), t.Day())
	if conf.Hourly {
		name = fmt.Sprintf("%v-%02d", name, t.Hour())
	}
	return
}

func setFile(t time.Time, prefix string) (f *os.File, filename string, err error) {
	name := logName(t, prefix)
	fname := filepath.Join(conf.LogDir, name)
	f, err = os.OpenFile(fname, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err == nil {
		symlink := filepath.Join(conf.LogDir, prefix)
		os.Remove(symlink)
		os.Symlink(name, symlink)
		return f, fname, nil
	}
	return nil, "", fmt.Errorf("log: cannot setFile log: %v", err)
}

func rotateTick(now time.Time) int {
	return (int(now.Unix()) + zoneOffset) / conf.rotateInterval
}

// createFile creates all the log files for severity from sev down to debugLog.
// l.mu is held.
func (l *loggingT) createFile(s uint) error {
	now := time.Now()
	idx, postfix := l.getFileDespInfo(s)
	sb := &syncBuffer{logger: l, sbName: conf.LogName + postfix, lastRotateTick: rotateTick(now)}
	if err := sb.rotateFile(now); err != nil {
		return err
	}
	l.file[idx] = sb
	return nil
}

const flushInterval = 500 * time.Millisecond

// flushDaemon periodically flushes the log file buffers.
func (l *loggingT) flushDaemon() {
	for range time.NewTicker(flushInterval).C {
		l.lockAndFlushAll()
	}
}

// lockAndFlushAll is like flushAll but locks l.mu first.
func (l *loggingT) lockAndFlushAll() {
	l.mu.Lock()
	l.flushAll()
	l.mu.Unlock()
}

// flushAll flushes all the logs and attempts to "sync" their data to disk.
// l.mu is held.
func (l *loggingT) flushAll() {
	for i := uint(0); i < maxLogCount; i++ {
		lfile := l.file[i]
		if lfile != nil && lfile.nlines > 0 {
			lfile.Flush()
			lfile.Sync()
			lfile.nlines = 0
		}
	}
}

func (l *loggingT) idleTimer() {
	if conf.GzipFileDays == 0 && conf.DelFileDays == 0 {
		return
	}

	fileTypes := []string{"log"}
	if conf.SepDbgLog {
		fileTypes = append(fileTypes, "dbg")
	}
	fileTypes = append(fileTypes, "clg")

	for {
		var sh time.Duration = 1
		now := time.Now()
		if now.Hour() >= 2 && now.Hour() <= 4 {
			sh = 20

			if conf.GzipFileDays > 0 {
				date := time.Now().Add(-time.Duration(conf.GzipFileDays*24) * time.Hour).Format("20060102")
				for _, tp := range fileTypes {
					src := fmt.Sprintf("%v.%v.%v", conf.LogName, tp, date)
					dest := src + ".gz"
					err := compressFile(dest, src)
					if err == nil {
						os.Remove(src)
					}
					Inf("压缩日志文件: %v -> %v, err:%v", src, dest, err)
				}
			}

			if conf.DelFileDays > 0 {
				date := time.Now().Add(-time.Duration(conf.GzipFileDays*24) * time.Hour).Format("20060102")
				for _, tp := range fileTypes {
					src := fmt.Sprintf("%v.%v.%v", conf.LogName, tp, date)
					err := os.Remove(src)
					Inf("删除日志文件:%v,err:%v", src, err)
				}
			}
		}
		time.Sleep(sh * time.Hour)
	}
}

func compressFile(Dst string, Src string) error {
	file, err := os.Open(Src)
	if err != nil {
		return err
	}

	newfile, err := os.Create(Dst)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zw := gzip.NewWriter(newfile)

	filestat, err := file.Stat()
	if err != nil {
		return err
	}

	zw.Name = filestat.Name()
	zw.ModTime = filestat.ModTime()
	_, err = io.Copy(zw, file)
	if err != nil {
		return err
	}

	zw.Flush()
	if err := zw.Close(); err != nil {
		return err
	}
	return nil
}

// debug 调试级别信息，生产环境不打印
func Dbg(format string, args ...interface{}) { logging.printf(debugLog, format, args...) }

// info 普通级别信息
func Inf(format string, args ...interface{}) { logging.printf(infoLog, format, args...) }

// warning 警告信息，程序出现小问题，需要加以重视
func Wrn(format string, args ...interface{}) { logging.printf(warningLog, format, args...) }

// error 错误信息，程序出错，需要马上处理
func Err(format string, args ...interface{}) { logging.printf(errorLog, format, args...) }

// fatal 致命错误，打印出堆栈，并退出程序，一般用在程序启动时进行检查，不可用于运行中的程序
func Ftl(format string, args ...interface{}) {
	flush()
	fmt.Fprintln(os.Stderr, "致命错误!!!", args)
	logging.printf(fatalLog, format, args...)
}

func Clog(format string, args ...interface{}) { logging.printf(clientLog, format, args...) }
