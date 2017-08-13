package RTJobRunner

type ParserObject interface {
	GetDependentJobs() []ParserObject
	GetJob() Request
	GetName() string
}
type CreateParserObjectFunc func() ParserObject





type JHJSONParserString struct {
	DependentJobs []*JHJSONParserString `json: "dependentJobs"`
	Job           string          `json: "job"`
	Name          string          `json: "name"`
	NumIterations int          `json: "num_iterations"`
}

func (this JHJSONParserString) GetDependentJobs() []*JHJSONParserString {
	return this.DependentJobs
}
func (this JHJSONParserString) GetJob() Request {
	return this.Job
}
func (this JHJSONParserString) GetName() string {
	return this.Name
}

func CreateJHJSONParserString() *JHJSONParserString {
	return &JHJSONParserString{}
}
