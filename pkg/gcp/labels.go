package gcp

import "strings"

type Kind int64

const (
	KindOnprem Kind = iota
	KindKNADA
	KindNAIS
	KindLegacy
	KindManagment
	KindUnknown
)

func ParseKind(in string) Kind {
	switch strings.ToLower(in) {
	case "knada":
		return KindKNADA
	case "onprem":
		return KindOnprem
	case "nais":
		return KindNAIS
	case "legacy":
		return KindLegacy
	case "managment":
		return KindManagment
	default:
		return KindUnknown
	}
}

func GetClusterServerForLegacyGCP(name string) string {
	switch name {
	case "prod-gcp":
		return "https://10.255.240.6"
	case "dev-gcp":
		return "https://10.255.240.5"
	case "ci-gcp":
		return "https://10.255.240.7"
	default:
		return "unknown-cluster"
	}
}
