package main

import (
	"fmt"
	"time"

	workpool "github.com/fox-wei/myworkpool"
)

func main() {

	p := workpool.New(5, workpool.WithPreAllcoWorker(false), workpool.WithBlock(false))

	time.Sleep(time.Second * 2)

	for i := 0; i < 10; i++ {
		err := p.Schedule(func() {
			time.Sleep(time.Second * 3)
			// fmt.Printf("woker[%03d]: working\n\n", i)
		})

		if err != nil {
			fmt.Printf("task[%d]: error:%s\n", i, err.Error())
		}
	}

	p.Free() //*释放资源
}
