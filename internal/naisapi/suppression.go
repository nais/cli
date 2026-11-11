package naisapi

const _ = `# @genqlient
	query MCPSuppressionFinding($id: ID!) {
		# @genqlient(bind: "encoding/json.RawMessage")
		node(id: $id) {
			__typename
			... on ImageVulnerability {
				id identifier package
				suppression { state }
			}
		}
	}
	mutation MCPApplySuppression($input: UpdateImageVulnerabilityInput!) {
		updateImageVulnerability(input: $input) {
			vulnerability {
				id
				suppression { state reason }
			}
		}
	}
`
