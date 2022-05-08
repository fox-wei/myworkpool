package workpool

import (
	"errors"
	"fmt"
	"sync"
)

type Task func() //^抽象请求

var (
	ErrNoIdleWorkerInPool = errors.New("no idle worker in pool") //! workerpool中任务已满，没有空闲goroutine用于处理新任务
	ErrWorkerPoolFreed    = errors.New("workerpool freed")       //! workerpool已终止运行
)

//?容量默认设置
const (
	defaultCapcity = 100
	maxCapcity     = 10000
)

type Pool struct {
	capacity int  //*pool容量
	preAllco bool //!是否创建pool的时候就创建workers，默认:false

	active chan struct{} //&计数器
	tasks  chan Task     //&task channel

	//*pool满的情况下，调用Sechdule是否阻塞当前goroutine，默认：true
	//*false，返回ErrNoWorkerAvailInPool
	block bool

	wg   sync.WaitGroup //~pool销毁等待worker退出
	quit chan struct{}  //~通知退出
}

func New(capacity int, opts ...Option) *Pool {
	if capacity <= 0 {
		capacity = defaultCapcity
	}

	if capacity > maxCapcity {
		capacity = maxCapcity
	}

	p := &Pool{
		capacity: capacity,
		block:    true,
		tasks:    make(chan Task),
		quit:     make(chan struct{}),
		active:   make(chan struct{}, capacity),
	}

	for _, opt := range opts {
		opt(p) //*执行option方法，为字段block, preAllco赋值
	}

	fmt.Printf("Workerpool start(preAllco=%t)\n", p.preAllco)

	//!预创建所有worker
	if p.preAllco {
		//*create all gorotines and into works channel
		for i := 0; i < p.capacity; i++ {
			p.newWorker(i + 1)
			p.active <- struct{}{}
		}
	}

	go p.run()

	return p
}

func (p *Pool) run() {
	idx := len(p.active)

	//*没有预创建worker，满员退出loop
	if !p.preAllco {
	loop:
		for t := range p.tasks {
			p.returnTask(t) //*p.tasks <- t
			select {
			case <-p.quit:
				return
			case p.active <- struct{}{}:
				idx++
				p.newWorker(idx)
			default:
				break loop
			}
		}
	}

	//*其它worker退出加入新的worker
	for {
		select {
		case <-p.quit:
			return
		case p.active <- struct{}{}:
			idx++
			p.newWorker(idx)
		}
	}
}

func (p *Pool) newWorker(i int) {
	p.wg.Add(1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("worker[%03d]: recover panic[%s] and exit\n", i, err)
				<-p.active
			}
			p.wg.Done()
		}()

		fmt.Printf("worker[%03d]: start\n", i)

		for {
			select {
			case <-p.quit:
				fmt.Printf("worker[%03d]: exit\n", i)
				<-p.active
				return
			case t := <-p.tasks:
				fmt.Printf("worker[%03d]: receive a task\n", i)
				t()
			}
		}
	}()

}

func (p *Pool) returnTask(t Task) {
	go func() {
		p.tasks <- t
	}()
}

func (p *Pool) Schedule(t Task) error {
	select {
	case <-p.quit:
		return ErrWorkerPoolFreed
	case p.tasks <- t:
		return nil
	default:
		//*是否继续阻塞或返回错误
		if p.block {
			p.tasks <- t
			return nil
		}
		return ErrNoIdleWorkerInPool
	}
}

func (p *Pool) Free() {
	close(p.quit) //*确保所有worker和p.run结束且Schedule返回nil
	p.wg.Wait()
	fmt.Printf("workerpool freed(preAlloc=%t)\n", p.preAllco)
}
