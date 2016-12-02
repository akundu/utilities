package jobs

type GoTaskResult struct {
	Status error
	Value  interface{}
}
type GoTask interface {
	Run() *GoTaskResult
}
