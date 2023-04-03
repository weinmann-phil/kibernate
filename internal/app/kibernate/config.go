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
	"regexp"
)

type WaitType string

const (
	WaitTypeConnect WaitType = "connect"
	WaitTypeLoading          = "loading"
	WaitTypeNone             = "none"
)

type Config struct {
	Namespace                     string
	Service                       string
	Deployment                    string
	ListenPort                    uint16
	ServicePort                   uint16
	IdleTimeoutSecs               uint16
	DefaultWaitType               WaitType
	ActivityPathMatch             *regexp.Regexp
	ActivityPathExclude           *regexp.Regexp
	WaitNonePathMatch             *regexp.Regexp
	WaitNonePathExclude           *regexp.Regexp
	WaitConnectPathMatch          *regexp.Regexp
	WaitConnectPathExclude        *regexp.Regexp
	WaitLoadingPathMatch          *regexp.Regexp
	WaitLoadingPathExclude        *regexp.Regexp
	UptimeMonitorUserAgentMatch   *regexp.Regexp
	UptimeMonitorUserAgentExclude *regexp.Regexp
	UptimeMonitorResponseCode     uint16
	UptimeMonitorResponseMessage  string
	NoDeactivationMoFrFromToUTC   []string
	NoDeactivationSatFromToUTC    []string
	NoDeactivationSunFromToUTC    []string
	NoDeactivationAutostart       bool
	ReadinessProbePath            string
	ReadinessTimeoutSecs          uint16
}
