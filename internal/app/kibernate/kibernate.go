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

import "log"

func NewKibernate(config Config) *Kibernate {
	return &Kibernate{config}
}

type Kibernate struct {
	Config Config
}

func (k *Kibernate) Run() error {
	log.Println("Starting kibernate")
	proxy, err := NewProxy(k.Config)
	if err != nil {
		log.Printf("Error creating proxy: %s", err.Error())
		return err
	}
	err = proxy.Start()
	if err != nil {
		log.Printf("Error starting proxy: %s", err.Error())
		return err
	}
	return nil
}
