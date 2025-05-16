package kubeconfig

import "strings"

type Kind int64

const (
	kindOnprem Kind = iota
	kindNAIS
	kindLegacy
	kindManagment
	kindUnknown
)

func parseKind(in string) Kind {
	switch strings.ToLower(in) {
	case "onprem":
		return kindOnprem
	case "nais":
		return kindNAIS
	case "legacy":
		return kindLegacy
	case "managment":
		return kindManagment
	default:
		return kindUnknown
	}
}

func getClusterServerForLegacyGCP(name string) string {
	switch name {
	case "ci-gcp":
		return "https://10.255.240.7"
	default:
		return "unknown-cluster"
	}
}
