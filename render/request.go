package render

type Request struct {
	Width    int
	Height   int
	BBOX     [4]float64
	EPSGCode int
	Format   string
}
