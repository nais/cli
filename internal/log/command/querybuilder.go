package command

import (
	"fmt"
	"strings"
)

type QueryBuilder struct {
	teams        []string
	environments []string
	workloads    []string
	containers   []string
}

// NewQueryBuilder creates a new instance of QueryBuilder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		teams:        make([]string, 0),
		environments: make([]string, 0),
		workloads:    make([]string, 0),
		containers:   make([]string, 0),
	}
}

// AddTeam adds a team in the service_namespace selector in the query.
func (qb *QueryBuilder) AddTeam(team string) *QueryBuilder {
	qb.teams = append(qb.teams, team)
	return qb
}

// AddEnvironment adds an environment in the k8s_cluster_name selector in the query.
func (qb *QueryBuilder) AddEnvironment(environment string) *QueryBuilder {
	qb.environments = append(qb.environments, environment)
	return qb
}

// AddWorkload adds a workload in the service_name selector in the query.
func (qb *QueryBuilder) AddWorkload(workload string) *QueryBuilder {
	qb.workloads = append(qb.workloads, workload)
	return qb
}

// AddContainer adds a container in the k8s_container_name filter in the query.
func (qb *QueryBuilder) AddContainer(container string) *QueryBuilder {
	qb.containers = append(qb.containers, container)
	return qb
}

// Build constructs the final query string that can be used to fetch logs.
func (qb *QueryBuilder) Build() string {
	selectors := []string{`service_name!=""`} // make sure we have at least one selector
	filters := make([]string, 0)

	if len(qb.teams) > 0 {
		selectors = append(selectors, fmt.Sprintf("service_namespace=~%q", strings.Join(qb.teams, "|")))
	}

	if len(qb.environments) > 0 {
		selectors = append(selectors, fmt.Sprintf("k8s_cluster_name=~%q", strings.Join(qb.environments, "|")))
	}

	if len(qb.workloads) > 0 {
		selectors = append(selectors, fmt.Sprintf("service_name=~%q", strings.Join(qb.workloads, "|")))
	}

	if len(qb.containers) > 0 {
		filters = append(filters, fmt.Sprintf("k8s_container_name=~%q", strings.Join(qb.containers, "|")))
	}

	query := fmt.Sprintf("{%s}", strings.Join(selectors, ","))

	for _, filter := range filters {
		query += " | " + filter
	}

	return query
}
