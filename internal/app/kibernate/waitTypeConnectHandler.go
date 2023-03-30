/*
   Copyright 2023 Michael Werner

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package kibernate

import (
	"log"
	"net/http"
)

type WaitTypeConnectHandler struct {
	Config     Config
	Proxy      *Proxy
	Deployment *DeploymentHandler
}

func NewWaitTypeConnectHandler(config Config, proxy *Proxy, deployment *DeploymentHandler) *WaitTypeConnectHandler {
	return &WaitTypeConnectHandler{
		Config:     config,
		Proxy:      proxy,
		Deployment: deployment,
	}
}

func (w *WaitTypeConnectHandler) Handle(writer http.ResponseWriter, request *http.Request) error {
	log.Printf("Handling request with wait type connect for path '%s' - waiting for deployment to become ready", request.URL.Path)
	w.Deployment.WaitForReady()
	log.Printf("Deployment is ready, proxying request for path '%s'", request.URL.Path)
	w.Proxy.PatchThrough(writer, request)
	return nil
}
