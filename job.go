package factory

type job struct {
	data   interface{}
	doneCh chan error
}
