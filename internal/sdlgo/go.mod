module gerrit.oran-osc.org/r/ric-plt/sdlgo

go 1.12

require (
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/stretchr/testify v1.3.0
)

replace gerrit.oran-osc.org/r/ric-plt/sdlgo/internal/sdlgoredis => ./internal/sdlgoredis
