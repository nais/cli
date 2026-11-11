package naisapi

const _ = `# @genqlient
	query MCPSuppressionPreview($target: VulnerabilitySuppressionTargetInput!) {
		vulnerabilitySuppressionPreview(target: $target) {
			...MCPSuppressionPreviewFields
		}
	}
	fragment MCPSuppressionPreviewFields on VulnerabilitySuppressionPreview {
		target {
			findingID advisoryID packagePurl workloadID team environment workload
			workloadType imageID imageName imageTag
		}
		affectedWorkloads { team environment workload workloadType imageTag }
		scopeFingerprint includesFutureMatches warning
	}
	mutation MCPApplySuppression($input: ApplyVulnerabilitySuppressionInput!) {
		applyVulnerabilitySuppression(input: $input) {
			persisted indeterminate state scopeFingerprint
		}
	}
`
