package RTJobRunner

import (
	"fmt"
	"bytes"
	"text/template"
	"os"
	"math/rand"
	"github.com/akundu/utilities/statistics/distribution"
	"regexp"
	"log"
)

type KV struct {
	key   string
	value int
}

/*
func print_path(path []*KV) string {
	var result string
	for i := range(path) {
		result += fmt.Sprintf("%s ", path[i].key)
	}
	return result
}
*/

type JSONJobProcessor struct {
	Name              	string
	CommandToExecute    string
	OriginalJSONJob     *JSONJob
}

func (this JSONJobProcessor) GetName() string {
	if len(this.Name) > 0 {
		return this.Name
	}
	return this.CommandToExecute
}

var template_expanders_regex *regexp.Regexp = nil
func add_jobs(jhjp *JSONJobContainer, json_jobs *JobHandler) error {
	json_job := jhjp.Job

	//create the results for each of the cases
	//1. expand the substitutes - add the job, if no substitution have happened
	f_results := ExpandSubstitutes(json_job.CommandToExecute, json_job.Substitutes)
	if(f_results == nil) { //didnt have anything to expand, so simply add the command to execute
		json_jobs.AddJob(NewRTRequestResultObject(
			&JSONJobProcessor{
				Name:			  jhjp.GetName(),
				CommandToExecute: json_job.CommandToExecute,
				OriginalJSONJob:  json_job,
			}))
		return nil
	}

	//2. template expand and start the job
	t, err := template.New(jhjp.GetName()).Parse(json_job.CommandToExecute)
	if err != nil {
		return err
	}
	lcw := bytes.NewBufferString("")
	for _, temp_result := range f_results {
		//t.Execute(lcw, f_results[i]) //expand the template
		t.Execute(lcw, temp_result)
		json_jobs.AddJob(NewRTRequestResultObject(
			&JSONJobProcessor{
				Name:			  jhjp.GetName(),
				CommandToExecute: lcw.String(),
				OriginalJSONJob:  json_job,
			}))

		lcw.Reset()
	}

	return nil
}


func ExpandSubstitutes(string_to_expand string, substitutes map[string]*SubstituteData) ([]map[string]string) {
	//1. get the list of keys
	res := template_expanders_regex.FindAllStringSubmatch(string_to_expand, -1)
	if len(res) == 0 { //didnt find anything to expand - simply add the job as it exists
		return nil
		return nil
	}
	keys := make([]string, len(res))
	for i, match := range(res) {
		keys[i] = match[1]
	}

	//2. initialize some params
	path := make([]*KV, 0, 10)
	f_results := make([]map[string]string, 0, 10)

	//3. recursively expand the substitutes
	expandSubstitutes(keys, 0, substitutes, path, &f_results)
	return f_results
}


func expandSubstitutes(keys []string,
	index int,
	substitutes map[string]*SubstituteData,
	path []*KV, //array of an object of
	result *[]map[string]string) {

	if index == len(keys) {
		//take the path till now and generate the data out of that
		obj_to_return := make(map[string]string)
		for _, path_obj := range path {
			kv_info := substitutes[path_obj.key]

			if kv_info.Type == "string" {
				obj_to_return[path_obj.key] = fmt.Sprintf("%s-%d", path_obj.key, path_obj.value)
			} else {
				obj_to_return[path_obj.key] = fmt.Sprintf("%d", path_obj.value)
			}
		}
		*result = append(*result, obj_to_return)
		return
	}


	kv_info, ok := substitutes[keys[index]]
	if(ok == false) {
		return
	}
	uniform_distr := distribution.NewuniformGenerator(kv_info.Lower, kv_info.Upper)
	gaussian_distr := distribution.NewgaussianGenerator(kv_info.Lower, kv_info.Upper, kv_info.NumToGenerate)

	for i := 0; i < kv_info.NumToGenerate; i++ {
		val := i
		if(kv_info.Type == "random") {
			val = uniform_distr.GenerateNumber()
		} else if (kv_info.Type == "gaussian") {
			val = gaussian_distr.GenerateNumber()
		}
		path = append(path, &KV{
			key:   keys[index],
			value: val,
		})

		expandSubstitutes(keys, index+1, substitutes, path, result)
		//pop from path
		path = path[:len(path)-1]
	}
	return
}

func init() {
	rand.Seed(int64(os.Getpid()))

	var err error
	if template_expanders_regex, err = regexp.Compile(`{{\.([0-9a-zA-Z]+)}}`); err != nil {
		log.Fatal("couldnt create regex obj for template_expanders")
	}
}
