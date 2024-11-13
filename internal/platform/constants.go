package platform

type Platform string

const (
	AWS Platform = "aws"
	GCP Platform = "gcp"
)

var AllPlatforms = []Platform{
	AWS,
	GCP,
}
