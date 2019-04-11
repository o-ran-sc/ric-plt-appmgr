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

package main

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -lmdclog
#
#include <mdclog/mdclog.h>
void xAppMgr_mdclog_write(mdclog_severity_t severity, const char *msg) {
     mdclog_write(severity, "%s", msg);
}
*/
import "C"

import (
	"fmt"
	"net/http"
	"time"
)

func mdclog(severity C.mdclog_severity_t, msg string) {
	msg = fmt.Sprintf("%s:: %s ", time.Now().Format("2019-01-02 15:04:05"), msg)

	C.mdclog_mdc_add(C.CString("XM"), C.CString("1.0.0"))
	C.xAppMgr_mdclog_write(severity, C.CString(msg))
}

func mdclogSetLevel(severity C.mdclog_severity_t) {
	C.mdclog_level_set(severity)
}

func Logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner.ServeHTTP(w, r)
		s := fmt.Sprintf("Logger: method=%s url=%s", r.Method, r.URL.RequestURI())
		mdclog(C.MDCLOG_DEBUG, s)
	})
}
