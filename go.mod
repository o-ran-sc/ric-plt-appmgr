module gerrit.oran-osc.org/r/ric-plt/appmgr

go 1.12

require (
	gerrit.oran-osc.org/r/ric-plt/sdlgo v0.0.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/gorilla/mux v1.7.1
	github.com/mitchellh/mapstructure v1.1.2
	github.com/orcaman/concurrent-map v0.0.0-20190314100340-2693aad1ed75
	github.com/segmentio/ksuid v1.0.2
	github.com/spf13/viper v1.3.2
	gopkg.in/yaml.v2 v2.2.2
)

replace gerrit.oran-osc.org/r/ric-plt/sdlgo => ./internal/sdlgo
