package factory

type state struct {
	callbacks []func(Job)
	jobCh     chan job
}
