package tunnel

// genqlient operation definitions used by go generate.

var _ = `# @genqlient
mutation CreateTunnel($input: CreateTunnelInput!) {
  createTunnel(input: $input) {
    tunnel {
      id
      name
      phase
      gatewayPublicKey
      forwarderEndpoint
      message
    }
  }
}
`

var _ = `# @genqlient
query GetTunnel($teamSlug: Slug!, $environmentName: String!, $name: String!) {
  team(slug: $teamSlug) {
    environment(name: $environmentName) {
      tunnel(name: $name) {
        id
        name
        phase
        gatewayPublicKey
        forwarderEndpoint
        message
      }
    }
  }
}
`

var _ = `# @genqlient
mutation DeleteTunnel($teamSlug: Slug!, $environmentName: String!, $tunnelName: String!) {
  deleteTunnel(input: { teamSlug: $teamSlug, environmentName: $environmentName, tunnelName: $tunnelName }) {
    success
  }
}
`
