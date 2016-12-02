package jobs

type GoTaskResult struct {
	Value  interface{}
	Status error
}
type GoTask interface {
	Run() *GoTaskResult
}
