package tunnel

// genqlient operation definitions — DO NOT REMOVE, used by go generate

var _ = `# @genqlient
mutation CreateTunnel($input: CreateTunnelInput!) {
  createTunnel(input: $input) {
    tunnel {
      id
      phase
      gatewayPublicKey
      gatewaySTUNEndpoint
      message
    }
  }
}
`

var _ = `# @genqlient
query GetTunnel($teamSlug: Slug!, $environmentName: String!, $id: ID!) {
  team(slug: $teamSlug) {
    environment(name: $environmentName) {
      tunnel(id: $id) {
        id
        phase
        gatewayPublicKey
        gatewaySTUNEndpoint
        message
      }
    }
  }
}
`

var _ = `# @genqlient
mutation UpdateTunnelSTUNEndpoint($tunnelID: ID!, $clientSTUNEndpoint: String!) {
  updateTunnelSTUNEndpoint(input: { tunnelID: $tunnelID, clientSTUNEndpoint: $clientSTUNEndpoint }) {
    tunnel {
      id
    }
  }
}
`

var _ = `# @genqlient
mutation DeleteTunnel($tunnelID: ID!) {
  deleteTunnel(input: { tunnelID: $tunnelID }) {
    success
  }
}
`
