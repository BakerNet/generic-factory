package factory

type job struct {
	data   Job
	doneCh chan error
}
