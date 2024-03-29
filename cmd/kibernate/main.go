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

package main

import (
	"flag"
	"github.com/kibernate/kibernate/internal/app/kibernate"
	"log"
	"regexp"
	"strings"
)

func main() {
	namespace := flag.String("namespace", "default", "The namespace of the service and deployment [default: default]")
	service := flag.String("service", "", "The name of the service to be proxied")
	deployment := flag.String("deployment", "", "The name of the deployment to be activated/deactivated")
	servicePort := flag.Uint("servicePort", 8080, "The port of the service to be proxied [default: 8080]")
	idleTimeoutSecs := flag.Uint("idleTimeoutSecs", 600, "The number of seconds to wait for activity before deactivating the deployment [default: 600]")
	defaultWaitType := flag.String("defaultWaitType", "connect", "The type of wait to perform by default - connect, loading, none [default: connect]")
	activityPathMatch := flag.String("activityPathMatch", ".*", "A regular expression to match paths that should be considered activity [default: \".*\"]")
	activityPathExclude := flag.String("activityPathExclude", "", "A regular expression to exclude paths that should not be considered activity")
	activityUserAgentMatch := flag.String("activityUserAgentMatch", ".*", "A regular expression to match User-Agent headers that should be considered activity [default: \".*\"]")
	activityUserAgentExclude := flag.String("activityUserAgentExclude", "", "A regular expression to exclude User-Agent headers that should not be considered activity")
	waitNonePathMatch := flag.String("waitNonePathMatch", "", "A regular expression to match paths that should not wait for deployment readiness")
	waitNonePathExclude := flag.String("waitNonePathExclude", "", "A regular expression to exclude paths that should not wait for deployment readiness")
	waitConnectPathMatch := flag.String("waitConnectPathMatch", "", "A regular expression to match paths that should wait for deployment readiness")
	waitConnectPathExclude := flag.String("waitConnectPathExclude", "", "A regular expression to exclude paths that should not wait for deployment readiness")
	waitLoadingPathMatch := flag.String("waitLoadingPathMatch", "", "A regular expression to match paths that should deliver a loading page while waiting for the deployment to be ready")
	waitLoadingPathExclude := flag.String("waitLoadingPathExclude", "", "A regular expression to exclude paths that should not deliver a loading page while waiting for the deployment to be ready")
	uptimeMonitorUserAgentMatch := flag.String("uptimeMonitorUserAgentMatch", "", "A regular expression to match User-Agent headers that should be considered uptime monitoring requests")
	uptimeMonitorUserAgentExclude := flag.String("uptimeMonitorUserAgentExclude", "", "A regular expression to exclude User-Agent headers that should not be considered uptime monitoring requests")
	uptimeMonitorResponseCode := flag.Uint("uptimeMonitorResponseCode", 200, "The HTTP response code to return for uptime monitoring requests [default: 200]")
	uptimeMonitorResponseMessage := flag.String("uptimeMonitorResponseMessage", "OK", "The HTTP response message to return for uptime monitoring requests [default: OK]")
	noDeactivationMoFrFromToUTC := flag.String("noDeactivationMoFrFromToUTC", "", "A from-to UTC time range in the format HH:MM-HH:MM that should not be considered for deactivation on Monday through Friday [default: none]")
	noDeactivationSatFromToUTC := flag.String("noDeactivationSatFromToUTC", "", "A from-to UTC time range in the format HH:MM-HH:MM that should not be considered for deactivation on Saturday [default: none]")
	noDeactivationSunFromToUTC := flag.String("noDeactivationSunFromToUTC", "", "A from-to UTC time range in the format HH:MM-HH:MM that should not be considered for deactivation on Sunday [default: none]")
	readinessProbePath := flag.String("readinessProbePath", "", "The path of the readiness probe [default: none]")
	readinessTimeoutSecs := flag.Uint("readinessTimeoutSecs", 30, "The number of seconds to wait for the readiness probe to return a 200 response before proxying requests anyway [default: 30]")
	noDeactivationAutostart := flag.Bool("noDeactivationAutostart", false, "If true, the deployment will autostart at the beginning of a configured no-deactivation time range [default: false]")
	flag.Parse()
	if *service == "" || *deployment == "" {
		panic("service and deployment must be set")
	}
	if *defaultWaitType != "connect" && *defaultWaitType != "loading" && *defaultWaitType != "none" {
		panic("defaultWaitType must be connect, loading, or none")
	}
	kibernateConfig := kibernate.Config{
		Namespace:                    *namespace,
		Service:                      *service,
		Deployment:                   *deployment,
		ServicePort:                  uint16(*servicePort),
		IdleTimeoutSecs:              uint16(*idleTimeoutSecs),
		DefaultWaitType:              kibernate.WaitType(*defaultWaitType),
		ListenPort:                   8080,
		UptimeMonitorResponseCode:    uint16(*uptimeMonitorResponseCode),
		UptimeMonitorResponseMessage: *uptimeMonitorResponseMessage,
		ReadinessTimeoutSecs:         uint16(*readinessTimeoutSecs),
		ReadinessProbePath:           *readinessProbePath,
		NoDeactivationAutostart:      *noDeactivationAutostart,
	}
	if *activityPathMatch != "" {
		kibernateConfig.ActivityPathMatch = regexp.MustCompile(*activityPathMatch)
	}
	if *activityPathExclude != "" {
		kibernateConfig.ActivityPathExclude = regexp.MustCompile(*activityPathExclude)
	}
	if *activityUserAgentMatch != "" {
		kibernateConfig.ActivityUserAgentMatch = regexp.MustCompile(*activityUserAgentMatch)
	}
	if *activityUserAgentExclude != "" {
		kibernateConfig.ActivityUserAgentExclude = regexp.MustCompile(*activityUserAgentExclude)
	}
	if *waitNonePathMatch != "" {
		kibernateConfig.WaitNonePathMatch = regexp.MustCompile(*waitNonePathMatch)
	}
	if *waitNonePathExclude != "" {
		kibernateConfig.WaitNonePathExclude = regexp.MustCompile(*waitNonePathExclude)
	}
	if *waitConnectPathMatch != "" {
		kibernateConfig.WaitConnectPathMatch = regexp.MustCompile(*waitConnectPathMatch)
	}
	if *waitConnectPathExclude != "" {
		kibernateConfig.WaitConnectPathExclude = regexp.MustCompile(*waitConnectPathExclude)
	}
	if *waitLoadingPathMatch != "" {
		kibernateConfig.WaitLoadingPathMatch = regexp.MustCompile(*waitLoadingPathMatch)
	}
	if *waitLoadingPathExclude != "" {
		kibernateConfig.WaitLoadingPathExclude = regexp.MustCompile(*waitLoadingPathExclude)
	}
	if *uptimeMonitorUserAgentMatch != "" {
		kibernateConfig.UptimeMonitorUserAgentMatch = regexp.MustCompile(*uptimeMonitorUserAgentMatch)
	}
	if *uptimeMonitorUserAgentExclude != "" {
		kibernateConfig.UptimeMonitorUserAgentExclude = regexp.MustCompile(*uptimeMonitorUserAgentExclude)
	}
	if *noDeactivationMoFrFromToUTC != "" {
		fromTo := strings.SplitN(*noDeactivationMoFrFromToUTC, "-", 2)
		if len(fromTo) != 2 {
			panic("noDeactivationMoFrFromToUTC must be in the format HH:MM-HH:MM")
		}
		if !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(fromTo[0]) || !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(fromTo[1]) {
			panic("noDeactivationMoFrFromToUTC must be in the format HH:MM-HH:MM")
		}
		kibernateConfig.NoDeactivationMoFrFromToUTC = fromTo
	}
	if *noDeactivationSatFromToUTC != "" {
		fromTo := strings.SplitN(*noDeactivationSatFromToUTC, "-", 2)
		if len(fromTo) != 2 {
			panic("noDeactivationSatFromToUTC must be in the format HH:MM-HH:MM")
		}
		if !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(fromTo[0]) || !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(fromTo[1]) {
			panic("noDeactivationSatFromToUTC must be in the format HH:MM-HH:MM")
		}
		kibernateConfig.NoDeactivationSatFromToUTC = fromTo
	}
	if *noDeactivationSunFromToUTC != "" {
		fromTo := strings.SplitN(*noDeactivationSunFromToUTC, "-", 2)
		if len(fromTo) != 2 {
			panic("noDeactivationSunFromToUTC must be in the format HH:MM-HH:MM")
		}
		if !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(fromTo[0]) || !regexp.MustCompile(`^\d\d:\d\d$`).MatchString(fromTo[1]) {
			panic("noDeactivationSunFromToUTC must be in the format HH:MM-HH:MM")
		}
		kibernateConfig.NoDeactivationSunFromToUTC = fromTo
	}
	kibernateInstance := kibernate.NewKibernate(kibernateConfig)
	err := kibernateInstance.Run()
	if err != nil {
		log.Fatalf("Error running kibernate: %s", err.Error())
	}
}
