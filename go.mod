module tlstunnel

go 1.22.2

require (
	github.com/google/uuid v1.6.0
	github.com/panlibin/gnet v1.0.6
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

require (
	github.com/libp2p/go-reuseport v0.0.1 // indirect
	github.com/panjf2000/ants/v2 v2.4.7 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/sys v0.24.0 // indirect
)

replace Wclient => ./Wclient
