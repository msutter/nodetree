package models

type StageTree struct {
	Description string
	ApiUser     string
	ApiPasswd   string
	Stages      []*Stage
}

func (st StageTree) GetStageByName(name string) (outStage *Stage) {
	for _, stage := range st.Stages {
		if stage.MatchName(name) {
			outStage = stage
		}
	}
	return outStage
}
