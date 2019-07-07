module testddosguard

go 1.12

require (
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	storage v0.0.0
	tcpserver v0.0.0
)

replace (
	storage => ../../packages/storage
	tcpserver => ../../packages/tcpserver
)
