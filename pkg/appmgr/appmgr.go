/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package appmgr

import (
	"flag"
	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

const DEFAULT_CONFIG_FILE = "../../config/appmgr.yaml"

var Logger *logger.Log

func LogRestRequests(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner.ServeHTTP(w, r)
		Logger.Info("Logger: method=%s url=%s", r.Method, r.URL.RequestURI())
	})
}

func parseCmd() string {
	var fileName *string
	fileName = flag.String("f", DEFAULT_CONFIG_FILE, "Specify the configuration file.")
	flag.Parse()

	return *fileName
}

func watch() {
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		Logger.Info("config file changed ", e.Name)
        setLoglevel()
	})
}

func setLoglevel() {
    var loglevel int
    viper.UnmarshalKey("loglevel", &loglevel)
    switch loglevel {
    case 1:
        Logger.Info("LOGLEVEL is set to ERROR\n")
    case 2:
        Logger.Info("LOGLEVEL is set to WARNING\n")
    case 3:
        Logger.Info("LOGLEVEL is set to INFO\n")
    case 4:
        Logger.Info("LOGLEVEL is set to DEBUG\n")
    }
    Logger.SetLevel(loglevel)
}
func loadConfig() {
	viper.SetConfigFile(parseCmd())
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	log.Printf("Using config file: %s\n", viper.ConfigFileUsed())

	// Watch for config file changes and re-read data ...
	watch()
}

func Init() {
	loadConfig()
	Logger = logger.NewLogger("appmgr")
    Logger.SetFormat(0)
    Logger.SetMdc("xm", "0.4.3")
    setLoglevel()
}
