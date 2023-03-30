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

type WaitTypeNoneHandler struct {
	Config Config
}

func NewWaitTypeNoneHandler(config Config) *WaitTypeNoneHandler {
	return &WaitTypeNoneHandler{
		Config: config,
	}
}

func (w *WaitTypeNoneHandler) Handle(writer http.ResponseWriter, request *http.Request) error {
	writer.WriteHeader(http.StatusServiceUnavailable)
	_, err := writer.Write([]byte("503 - Service Unavailable"))
	if err != nil {
		log.Printf("Error writing response: %s", err.Error())
		return err
	}
	return nil
}
