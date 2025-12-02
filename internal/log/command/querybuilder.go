package command

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	environments []string
	teams        []string
	workloads    []string
	containers   []string
	pods         []string
}

// NewQueryBuilder creates a new instance of QueryBuilder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		environments: make([]string, 0),
		teams:        make([]string, 0),
		workloads:    make([]string, 0),
		containers:   make([]string, 0),
		pods:         make([]string, 0),
	}
}

// AddEnvironments adds environments in the k8s_cluster_name selector in the query.
func (qb *QueryBuilder) AddEnvironments(environment ...string) *QueryBuilder {
	qb.environments = append(qb.environments, environment...)
	return qb
}

// AddTeams adds teams in the service_namespace selector in the query.
func (qb *QueryBuilder) AddTeams(team ...string) *QueryBuilder {
	qb.teams = append(qb.teams, team...)
	return qb
}

// AddWorkloads adds workloads in the service_name selector in the query.
func (qb *QueryBuilder) AddWorkloads(workload ...string) *QueryBuilder {
	qb.workloads = append(qb.workloads, workload...)
	return qb
}

// AddContainers adds containers in the k8s_container_name filter in the query.
func (qb *QueryBuilder) AddContainers(container ...string) *QueryBuilder {
	qb.containers = append(qb.containers, container...)
	return qb
}

// AddContainers adds containers in the k8s_container_name filter in the query.
func (qb *QueryBuilder) AddPods(pod ...string) *QueryBuilder {
	qb.pods = append(qb.pods, pod...)
	return qb
}

// Build constructs the final query string that can be used to fetch logs.
func (qb *QueryBuilder) Build() string {
	selectors := []string{`service_name!=""`} // make sure we have at least one selector
	filters := make([]string, 0)

	if len(qb.environments) > 0 {
		selectors = append(selectors, fmt.Sprintf("k8s_cluster_name=~%q", strings.Join(qb.environments, "|")))
	}

	if len(qb.teams) > 0 {
		selectors = append(selectors, fmt.Sprintf("service_namespace=~%q", strings.Join(qb.teams, "|")))
	}

	if len(qb.workloads) > 0 {
		selectors = append(selectors, fmt.Sprintf("service_name=~%q", strings.Join(qb.workloads, "|")))
	}

	if len(qb.containers) > 0 {
		filters = append(filters, fmt.Sprintf("k8s_container_name=~%q", strings.Join(qb.containers, "|")))
	}

	if len(qb.pods) > 0 {
		filters = append(filters, fmt.Sprintf("k8s_pod_name=~%q", strings.Join(qb.pods, "|")))
	}

	query := fmt.Sprintf("{%s}", strings.Join(selectors, ","))

	for _, filter := range filters {
		query += " | " + filter
	}

	return query
}
