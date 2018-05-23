package api

import (
	"github.com/SafeScale/providers"
	providerapi "github.com/SafeScale/providers/api"
	"github.com/SafeScale/providers/api/VMState"

	"github.com/SafeScale/cluster/api/ClusterState"
	"github.com/SafeScale/cluster/api/Complexity"
	"github.com/SafeScale/cluster/api/Flavor"
	"github.com/SafeScale/cluster/api/NodeType"
)

const (
	//ClusterContainerName contains the name of the Object Storage container where to put Cluster definitions
	DeployContainerName = "0.deploy"

	// Prefix to use to complete pathes of cluster definitions
	ClusterContainerNamePrefix = "0.clusters/"
)

//NodeRequest defines what kind of node is wanted
type NodeRequest struct {
	//Size contains the requested sizing of the node
	Size     providerapi.VMSize
	Template providerapi.VMTemplate
	Type     NodeType.Enum
}

//Node represents a created Node
type Node struct {
	//ID is the unique identifier of the VM of the node
	ID string
	//Template
	TemplateID string
	//Type
	Type NodeType.Enum
	//State
	State VMState.Enum
}

//ClusterRequest defines what kind of Cluster is wanted
type ClusterRequest struct {
	//Name is the name of the cluster wanted
	Name string

	//CIDR defines the network to create
	CIDR string

	//Mode is the implementation wanted, can be Simple, HighAvailability or HighVolume
	Complexity Complexity.Enum
}

//ClusterAPI is an interface of methods associated to Cluster-like structs
type ClusterAPI interface {
	//Start starts the cluster
	Start() error
	//Stop stops the cluster
	Stop() error
	//GetState returns the current state of the cluster
	GetState() (ClusterState.Enum, error)
	//AddNode adds a node
	AddNode(NodeType.Enum, providerapi.VMRequest) (*Node, error)
	//DeleteNode deletes a node
	DeleteNode(string) error
	//ListNodes lists the nodes in the cluster
	ListNodes() (*[]Node, error)
	//getNode returns a node based on its ID
	GetNode(string) (*Node, error)

	//SaveClusterDefinition
	SaveDefinition() error
	//ReadClusterDefinition
	ReadDefinition() (bool, error)
	//RemoveClusterDefinition
	RemoveDefinition() error
}

//Cluster contains the bare minimum information about a cluster
type Cluster struct {
	//Name is the name of the cluster
	Name string

	//CIDR is the network CIDR wanted for the Network
	CIDR string

	//Mode is the mode of cluster; can be Simple, HighAvailability, HighVolume
	Complexity Complexity.Enum

	//Keypair contains the key-pair used inside the Cluster
	Keypair *providerapi.KeyPair

	//State
	State ClusterState.Enum
}

//ClusterManager contains the bare minimum of information about a cluster manager
// A Manager is able to handle many clusters on the same tenant and of the same flavor.
type ClusterManager struct {
	//Service is the provider service used to managed infrastructure
	Service *providers.Service

	//Tenant where is hosted the infrastructure of the cluster
	Tenant string

	//Flavor is the name of the cluster method (currently can be only DCOS)
	Flavor Flavor.Enum
}

//ClusterManagerAPI is an interface of methods associated to Manager-like structs
type ClusterManagerAPI interface {
	//GetService returns the service client of the tenant
	GetService() *providers.Service
	//GetTenantName returns the name of the tenant used by the manager
	GetTenantName() string

	//CreateNetwork creates a network named name
	CreateCluster(req ClusterRequest) (ClusterAPI, error)
	//ListClusters lists the clusters availables
	ListClusters() (*[]string, error)
	//GetCluster returns Cluster instance corresponding to the cluster named 'name'
	GetCluster(name string) (ClusterAPI, error)
	//DeleteCluster deletes a cluster identified by its name
	DeleteCluster(name string) error

	//StartCluster starts a cluster identified by its name
	StartCluster(name string) error
	//StopCluster stops a cluster identified by its name
	StopCluster(name string) error
}