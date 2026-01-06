package redis

type ClientType string

const (
	ClientTypeSingleNode ClientType = "single_node"
	ClientTypeCluster    ClientType = "cluster"
)

type Config struct {
	URL        string
	DB         int
	Password   string
	ClientType ClientType
}
