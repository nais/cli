package naisapi

const _ = `# @genqlient
	query MCPApplicationVulnerabilities($team: Slug!, $name: String!, $env: String!, $first: Int!, $after: Cursor) {
		team(slug: $team) {
			slug
			environment(name: $env) {
				application(name: $name) {
					id name
					teamEnvironment { environment { name } }
					ingresses { url type }
					authIntegrations { __typename }
					networkPolicy { ...MCPVulnerabilityNetwork }
					image { ...MCPVulnerabilityImage }
				}
			}
		}
	}
	query MCPJobVulnerabilities($team: Slug!, $name: String!, $env: String!, $first: Int!, $after: Cursor) {
		team(slug: $team) {
			slug
			environment(name: $env) {
				job(name: $name) {
					id name
					teamEnvironment { environment { name } }
					authIntegrations { __typename }
					networkPolicy { ...MCPVulnerabilityNetwork }
					image { ...MCPVulnerabilityImage }
				}
			}
		}
	}
	fragment MCPVulnerabilityNetwork on NetworkPolicy {
		inbound { rules { targetWorkloadName targetTeamSlug mutual } }
	}
	fragment MCPVulnerabilityImage on ContainerImage {
		id name tag digest
		sbom { status }
		vulnerabilitySummary { total lastUpdated }
		vulnerabilities(first: $first, after: $after, orderBy: {field: IDENTIFIER, direction: ASC}) {
			pageInfo { hasNextPage endCursor }
			nodes {
				id identifier severity package description
				riskAssessment { cvssScore epssScore epssPercentile hasKevEntry knownRansomwareUse }
				remediation { fixVersion latestVersion }
				suppression { state }
				vulnerabilityDetailsLink
			}
		}
	}
`
