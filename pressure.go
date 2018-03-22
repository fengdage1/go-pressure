package main
import(
	"fmt"
	"sync"
	"net/http"
	"sync/atomic"
	"io/ioutil"
	"time"
	"os"
	"strconv"
)

func NewWorker(total,concurrency int64,url string,timeout time.Duration,errout bool) *worker{
	wk := new(worker)
	wk.total=total
	wk.concurrency=concurrency
	wk.url=url
	wk.job_queue = make(chan struct{},concurrency)
	wk.timeout=timeout
	wk.errout=errout
	return wk
}

type worker struct{
	job_queue chan struct{}
	wg sync.WaitGroup
	total int64
	concurrency int64
	url string
	failed int64
	sucessed int64
	recved int64
	timeout time.Duration
	errout bool
	starttime time.Time
}


func (this *worker) check(){
	fmt.Println("\n\n---------------------------------")
	fmt.Println("result:")
	fmt.Println("---------------------------------")
	duration:=int64(time.Since(this.starttime).Seconds())+1
	fmt.Println(fmt.Sprintf("timeused:%d\ncomplete:%d\nfailed:%d\nsuccesspercent:%d%%\nrecved:%d\nresp/sec:%d\n",duration,this.failed+this.sucessed,this.failed,this.sucessed*100/this.total,this.recved,int64(this.recved/duration)))
}


func (this *worker) run(){
	var i,done,subdone int64
	this.starttime = time.Now()
	sub:=this.total/10
	index := 0
	fmt.Println("start...")

	for i=0;i<this.total;i++{
		this.wg.Add(1)
		this.job_queue<-struct{}{}
		subdone++
		if subdone>=sub {
			index++
			if index < 10{
				done += subdone
				subdone = 0
				duration:=int64(time.Since(this.starttime).Seconds())+1
				fmt.Println("------------------------")
				fmt.Println(fmt.Sprintf("done:%d\ntimeused:%d\nfailed:%d\nrecved:%d\nresp/sec:%d", done, duration,this.failed, this.recved, int64(this.recved/int64(time.Since(this.starttime).Seconds())+1)))
			}
		}
	}
}

func (this *worker) wait(){
	this.wg.Wait()
}

func(this *worker) initThread(){
	var i int64
	for i=0;i<this.concurrency;i++{
		go this.thread()
	}
}

func (this *worker) thread(){
	client:=&http.Client{
		Timeout: this.timeout,
	}
	for range this.job_queue{
		res,err:=client.Get(this.url)
		if err != nil{
			if this.errout{
				fmt.Println(err)
			}
			atomic.AddInt64(&this.failed,1)
			this.wg.Done()
			continue
		}
		if res.StatusCode != 200 && res.StatusCode != 304{
			if this.errout{
				fmt.Println(err)
			}
			atomic.AddInt64(&this.failed,1)
			res.Body.Close()
			this.wg.Done()
			continue
		}
		html,err:=ioutil.ReadAll(res.Body)
		if err != nil{
			if this.errout{
				fmt.Println(err)
			}
			atomic.AddInt64(&this.failed,1)
			res.Body.Close()
			this.wg.Done()
			continue
		}
		length := len(html)
		atomic.AddInt64(&this.recved,int64(length))
		atomic.AddInt64(&this.sucessed,1)
		//fmt.Println(id,this.recved,this.failed,this.sucessed)
		res.Body.Close()
		this.wg.Done()
	}
}

func errargs(){
	fmt.Println("error too few args\n-n total\n-c concurrency\n-u url")
	fmt.Println("extend args:\n-t timeout(second,default 5)\n-e errout")
	fmt.Println("example1: pressure -n 10000 -c 1000 -u http://www.google.com")
	fmt.Println("example2: pressure -n 10000 -c 1000 -u http://www.google.com -t 10 -e")
}

func main(){
	if len(os.Args)<2{
		errargs()
		os.Exit(1)
	}
	var total,concurrency int64
	var errout bool
	var timeout time.Duration = time.Second*5
	var url string
	for i:=1;i<len(os.Args);i++{
		switch(os.Args[i]) {
		case "-n":
			t, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			total = int64(t)
			i++
		case "-c":
			c, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			concurrency = int64(c)
			i++
		case "-u":
			url = os.Args[i+1]
			i++
		case "-t":
			t,err:=strconv.Atoi(os.Args[i+1])
			if err != nil{
				fmt.Println(err)
				os.Exit(-1)
			}
			timeout=time.Duration(t) * time.Second
			i++
		case "-e":
			errout = true
		default:
			errargs()
			os.Exit(-1)
		}
	}
	if total==0||concurrency==0||url==""{
		errargs()
		os.Exit(-1)
	}
	wk := NewWorker(total,concurrency,url,timeout,errout)
	wk.initThread()
	wk.run()
	wk.wait()
	wk.check()
	fmt.Println("finish")
}
