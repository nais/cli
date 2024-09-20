package kubeconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
)

type k8sCluster struct {
	Name        string
	Endpoint    string
	Location    string
	CA          string
	Tenant      string
	User        *onpremUser
	Kind        Kind
	Environment string
}

type onpremUser struct {
	ServerID string `json:"serverID"`
	ClientID string `json:"clientID"`
	TenantID string `json:"tenantID"`
	UserName string `json:"userName"`
}

func getClustersFromGCP(ctx context.Context, tenant string, options filterOptions) ([]k8sCluster, error) {
	projects, err := getProjects(ctx, tenant, options)
	if err != nil {
		return nil, err
	}

	clusters, err := getClusters(ctx, projects, options)
	if err != nil {
		return nil, err
	}

	if options.prefixWithTenant {
		for i, cluster := range clusters {
			if options.skipNAVPrefix && cluster.Tenant == "nav" {
				continue
			}

			cluster.Name = cluster.Tenant + "-" + strings.TrimPrefix(cluster.Name, "nais-")
			clusters[i] = cluster
		}
	}

	return clusters, nil
}

func getClusters(ctx context.Context, projects []project, options filterOptions) ([]k8sCluster, error) {
	var clusters []k8sCluster
	for _, project := range projects {
		if options.verbose {
			fmt.Println("getting clusters for", project.ID)
		}
		var cluster []k8sCluster
		var err error

		switch project.Kind {
		case kindOnprem:
			cluster, err = getOnpremClusters(ctx, project)
		default:
			cluster, err = getGCPClusters(ctx, project)
		}

		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster...)
	}

	return clusters, nil
}

func getGCPClusters(ctx context.Context, project project) ([]k8sCluster, error) {
	svc, err := container.NewService(ctx)
	if err != nil {
		return nil, err
	}

	call := svc.Projects.Locations.Clusters.List("projects/" + project.ID + "/locations/-")
	response, err := call.Do()
	if err != nil {
		return nil, err
	}

	var clusters []k8sCluster
	for _, cluster := range response.Clusters {
		name := cluster.Name
		if cluster.Name == "knada-gke" {
			name = "knada"
		}

		kind := project.Kind
		if slices.Contains([]string{"dev-gcp", "prod-gcp", "ci-gcp"}, name) {
			kind = kindLegacy
		}

		clusters = append(clusters, k8sCluster{
			Name:        name,
			Endpoint:    "https://" + cluster.Endpoint,
			Location:    cluster.Location,
			CA:          cluster.MasterAuth.ClusterCaCertificate,
			Tenant:      project.Tenant,
			Kind:        kind,
			Environment: project.Name,
		})
	}
	return clusters, nil
}

func getOnpremClusters(ctx context.Context, project project) ([]k8sCluster, error) {
	if project.Kind != kindOnprem {
		return nil, nil
	}

	svc, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}
	proj, err := svc.Projects.Get(project.ID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	var clusters []k8sCluster
	for _, meta := range proj.CommonInstanceMetadata.Items {
		if meta.Key != "kubeconfig" || meta.Value == nil {
			continue
		}

		config := &struct {
			ServerID string `json:"serverID"`
			ClientID string `json:"clientID"`
			TenantID string `json:"tenantID"`
			URL      string `json:"url"`
			UserName string `json:"userName"`
		}{}
		if err := json.Unmarshal([]byte(*meta.Value), &config); err != nil {
			return nil, err
		}

		clusters = append(clusters, k8sCluster{
			Name:     project.Name,
			Endpoint: config.URL,
			Tenant:   "nav",
			Kind:     kindOnprem,
			User: &onpremUser{
				ServerID: config.ServerID,
				ClientID: config.ClientID,
				TenantID: config.TenantID,
				UserName: config.UserName,
			},
		})

		return clusters, nil

	}

	return clusters, nil
}
