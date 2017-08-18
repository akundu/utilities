package RTJobRunner

type ParserObject interface {
	GetPostJobs() []ParserObject
	GetJob() Request
	GetName() string
}
type CreateParserObjectFunc func() ParserObject





type JHJSONParserString struct {
	PostJobs []*JHJSONParserString `json: "postJobs"`
	PreJobs []*JHJSONParserString `json: "preJobs"`
	Job           string          `json: "job"`
	Name          string          `json: "name"`
	NumIterations int          `json: "numIterations"`
}

func (this JHJSONParserString) GetPostJobs() []*JHJSONParserString {
	return this.PostJobs
}
func (this JHJSONParserString) GetPreJobs() []*JHJSONParserString {
	return this.PreJobs
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
