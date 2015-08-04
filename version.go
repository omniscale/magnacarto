package magnacarto

var Version string

// buidVersion gets replaced while building with
// go build -ldflags "-X github.com/omniscale/magnacarto.buildVersion 1234"
var buildVersion string

func init() {
	Version = "0.1"
	Version += buildVersion
}
