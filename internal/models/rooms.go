package models

type AllocateRequest struct {
	Semester string `json:"semester"`
	Faculty  string `json:"faculty"`

	Programs []struct {
		Name         string `json:"name"`
		Classrooms   int    `json:"classrooms"`
		Laboratories int    `json:"laboratories"`
	} `json:"programs"`
}

type ProgramAllocation struct {
	Name         string   `json:"name"`
	Classrooms   []string `json:"classrooms"`
	Laboratories []string `json:"laboratories"`
	Adapted      []string `json:"adapted"`
}

type AllocateResponse struct {
	Semester string `json:"semester"`
	Faculty  string `json:"faculty"`

	Programs []ProgramAllocation `json:"programs"`
}
