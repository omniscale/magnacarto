package mml

type Map struct {
	ZoomScales []int  `yaml:"ZoomScales"`
	SRS        string `yaml:"SRS"`
	BBOX       []int  `yaml:"BBOX"`
}
