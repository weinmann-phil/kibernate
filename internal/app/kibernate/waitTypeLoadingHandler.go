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
	"context"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
)

type WaitTypeLoadingHandler struct {
	Config      Config
	LoadingHtml string
}

func NewWaitTypeLoadingHandler(config Config) (*WaitTypeLoadingHandler, error) {
	clientConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("Error creating in-cluster config: %s", err.Error())
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		log.Printf("Error creating client set: %s", err.Error())
		return nil, err
	}
	loadingHtmlConfigMap, err := clientSet.CoreV1().ConfigMaps(config.Namespace).Get(context.TODO(), "kibernate-loading-html", metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting kibernate-loading-html config map: %s", err.Error())
		return nil, err
	}
	loadingHtml, ok := loadingHtmlConfigMap.Data["loading.html"]
	if !ok {
		log.Println("loading.html not found in kibernate-loading-html config map")
		return nil, errors.New("loading.html not found in kibernate-loading-html config map")
	}
	return &WaitTypeLoadingHandler{
		Config:      config,
		LoadingHtml: loadingHtml,
	}, nil
}

func (w *WaitTypeLoadingHandler) Handle(writer http.ResponseWriter, request *http.Request) error {
	writer.Header().Set("Content-Type", "text/html")
	writer.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	writer.Header().Set("Pragma", "no-cache")
	writer.Header().Set("Expires", "0")
	_, err := writer.Write([]byte(w.LoadingHtml))
	return err
}
