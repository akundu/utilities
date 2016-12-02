package jobs

import (
	"sync"
	"time"

	"github.com/akundu/utilities/containers"
)

type GoJob struct {
	tasks []GoTask
}

//pair is of type Pair{[]*Pair{status, time_taken}, total time taken to complete the job}
func (this *GoJob) Run() ([]*containers.Tuple, time.Duration) {
	results_saved := make([]*containers.Tuple, 0, len(this.tasks))
	start_time_overall := time.Now()
	var wg sync.WaitGroup
	for task_index, task := range this.tasks {
		local_task_index := task_index
		wg.Add(1)
		go func() {
			defer wg.Done()
			var delta time.Duration = 0
			start_time := time.Now()
			result := task.Run()
			if result.Status == nil {
				delta = time.Since(start_time)
			}

			results_saved[local_task_index] = containers.NewTuple(&result, delta)
		}()
	}
	wg.Wait()
	delta_time_overall := time.Since(start_time_overall)

	return results_saved, delta_time_overall
}
func (this *GoJob) AddTask(gt GoTask) {
	this.tasks = append(this.tasks, gt)
}

func newGoJob() *GoJob {
	job := GoJob{make([]GoTask, 0, 5)}
	return &job
}
