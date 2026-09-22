package constants

type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderLocal  Provider = "local"
)

func (p Provider) IsValid() bool {
	switch p {
	case ProviderGoogle, ProviderLocal:
		return true
	default:
		return false
	}
}
