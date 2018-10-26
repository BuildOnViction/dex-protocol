package common

import (
	"os"
)

const (
	BzzDefaultNetworkId = 4242
	WSDefaultPort       = 18543
	BzzDefaultPort      = 8542
	P2pPort             = 30100
	IPCName             = "demo.ipc"
	DatadirPrefix       = ".data_"
)

var (
	basePath, _ = os.Getwd()
)
