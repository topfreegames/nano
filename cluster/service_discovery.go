package cluster

// ServiceDiscoveryClient is the interface for a service discovery client
type ServiceDiscoveryClient interface {
	Register(serverType, serverID, serverInfo string) error
	GetServers() (map[string]map[string]string, error)
	GetServersOfType(serverType string) (map[string]string, error)
	GetServerTypes() ([]string, error)
	UpdateServerInfo(serverID string) error
}
