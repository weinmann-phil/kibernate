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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type Proxy struct {
	Config                 Config
	TargetBaseUrl          *url.URL
	HttpServer             *http.Server
	WaitTypeNoneHandler    WaitTypeHandler
	WaitTypeConnectHandler WaitTypeHandler
	WaitTypeLoadingHandler WaitTypeHandler
	DefaultWaitTypeHandler WaitTypeHandler
	LastActivity           time.Time
	Deployment             *DeploymentHandler
}

func NewProxy(config Config) (*Proxy, error) {
	targetBaseUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", config.Service, config.ServicePort))
	if err != nil {
		return nil, err
	}
	httpServer := http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", config.ListenPort),
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	p := &Proxy{Config: config, TargetBaseUrl: targetBaseUrl, HttpServer: &httpServer}
	p.HttpServer.Handler = p
	p.Deployment, err = NewDeploymentHandler(p.Config)
	if err != nil {
		return nil, err
	}
	p.WaitTypeConnectHandler = NewWaitTypeConnectHandler(p.Config, p, p.Deployment)
	p.WaitTypeLoadingHandler, err = NewWaitTypeLoadingHandler(p.Config)
	if err != nil {
		return nil, err
	}
	p.WaitTypeNoneHandler = NewWaitTypeNoneHandler(p.Config)
	switch p.Config.DefaultWaitType {
	case WaitTypeConnect:
		p.DefaultWaitTypeHandler = p.WaitTypeConnectHandler
	case WaitTypeLoading:
		p.DefaultWaitTypeHandler = p.WaitTypeLoadingHandler
	case WaitTypeNone:
		p.DefaultWaitTypeHandler = p.WaitTypeNoneHandler
	}
	return p, nil
}

func (p *Proxy) Start() error {
	go func() {
		err := p.ContinuouslyCheckIdleness()
		if err != nil {
			panic(err.Error())
		}
	}()
	return p.HttpServer.ListenAndServe()
}

func (p *Proxy) ContinuouslyCheckIdleness() error {
	for range time.Tick(10 * time.Second) {
		if time.Since(p.LastActivity) > time.Duration(p.Config.IdleTimeoutSecs)*time.Second && p.Deployment.Status == DeploymentStatusReady && time.Since(p.Deployment.LastStatusChange) > time.Duration(p.Config.IdleTimeoutSecs)*time.Second {
			err := p.Deployment.DeactivateDeployment()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Proxy) PatchThrough(writer http.ResponseWriter, request *http.Request) {
	originalUrl := request.URL
	request.URL.Host = p.TargetBaseUrl.Host
	request.URL.Scheme = p.TargetBaseUrl.Scheme
	request.URL.Path = fmt.Sprintf("%s%s", p.TargetBaseUrl.Path, originalUrl.Path)
	request.URL.RawPath = fmt.Sprintf("%s%s", p.TargetBaseUrl.Path, originalUrl.RawPath)
	if p.Config.CustomHostValue != "" {
		request.Host = p.Config.CustomHostValue
	} else {
		request.Host = p.TargetBaseUrl.Host
	}
	request.RequestURI = ""
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	for key, values := range response.Header {
		for _, value := range values {
			writer.Header().Add(key, value)
		}
	}
	writer.WriteHeader(response.StatusCode)
	_, err = io.Copy(writer, response.Body)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *Proxy) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var err error
	if p.Config.UptimeMonitorUserAgentMatch != nil && p.Config.UptimeMonitorUserAgentMatch.MatchString(request.Header.Get("User-Agent")) {
		if p.Config.UptimeMonitorUserAgentExclude == nil || !p.Config.UptimeMonitorUserAgentExclude.MatchString(request.Header.Get("User-Agent")) {
			if p.Deployment.Status == DeploymentStatusReady {
				p.PatchThrough(writer, request)
			} else {
				writer.Header().Add("Content-Type", "text/plain")
				writer.WriteHeader(http.StatusOK)
				_, err = writer.Write([]byte(p.Config.UptimeMonitorResponseMessage))
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
				}
			}
			return
		}
	}
	if p.IsPathConsideredActivity(request.URL.Path) {
		p.LastActivity = time.Now()
	}
	if p.Deployment.Status == DeploymentStatusReady {
		p.PatchThrough(writer, request)
	} else {
		err = p.Deployment.ActivateDeployment()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if p.IsPathMatchingFor(WaitTypeConnect, request.URL.Path) {
			err = p.WaitTypeConnectHandler.Handle(writer, request)
		} else if p.IsPathMatchingFor(WaitTypeLoading, request.URL.Path) {
			err = p.WaitTypeLoadingHandler.Handle(writer, request)
		} else if p.IsPathMatchingFor(WaitTypeNone, request.URL.Path) {
			err = p.WaitTypeNoneHandler.Handle(writer, request)
		} else {
			err = p.DefaultWaitTypeHandler.Handle(writer, request)
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (p *Proxy) Stop() error {
	return p.HttpServer.Close()
}

func (p *Proxy) IsPathConsideredActivity(path string) bool {
	pathMatch := p.Config.ActivityPathMatch
	pathExclude := p.Config.ActivityPathExclude
	isMatching := false
	if pathMatch != nil && pathMatch.MatchString(path) {
		isMatching = true
		if pathExclude != nil && pathExclude.MatchString(path) {
			isMatching = false
		}
	}
	return isMatching
}

func (p *Proxy) IsPathMatchingFor(waitType WaitType, path string) bool {
	var pathMatch *regexp.Regexp
	var pathExclude *regexp.Regexp
	isMatching := false
	switch waitType {
	case WaitTypeConnect:
		pathMatch = p.Config.WaitConnectPathMatch
		pathExclude = p.Config.WaitConnectPathExclude
	case WaitTypeLoading:
		pathMatch = p.Config.WaitLoadingPathMatch
		pathExclude = p.Config.WaitLoadingPathExclude
	case WaitTypeNone:
		pathMatch = p.Config.WaitNonePathMatch
		pathExclude = p.Config.WaitNonePathExclude
	}
	if pathMatch != nil && pathMatch.MatchString(path) {
		isMatching = true
		if pathExclude != nil && pathExclude.MatchString(path) {
			isMatching = false
		}
	}
	return isMatching
}
