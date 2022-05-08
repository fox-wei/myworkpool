package workpool

type Option func(*Pool)

//?Schedule调用是否阻塞
func WithBlock(block bool) Option {
	return func(p *Pool) {
		p.block = block
	}
}

//?是否预创建所有worker
func WithPreAllcoWorker(preAllco bool) Option {
	return func(p *Pool) {
		p.preAllco = preAllco
	}
}
