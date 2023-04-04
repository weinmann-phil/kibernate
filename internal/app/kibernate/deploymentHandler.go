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
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"net/url"
	"time"
)

type DeploymentStatus string

const (
	DeploymentStatusReady        DeploymentStatus = "ready"
	DeploymentStatusActivating                    = "activating"
	DeploymentStatusDeactivating                  = "deactivating"
	DeploymenStatusDeactivated                    = "deactivated"
)

type DeploymentHandler struct {
	Config           Config
	Status           DeploymentStatus
	LastStatusChange time.Time
	KubeClientSet    *kubernetes.Clientset
}

func NewDeploymentHandler(config Config) (*DeploymentHandler, error) {
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
	d := &DeploymentHandler{Config: config, KubeClientSet: clientSet}
	err = d.UpdateStatus(nil)
	if err != nil {
		log.Printf("Error updating deployment status: %s", err.Error())
		return nil, err
	}
	go func() {
		for {
			err := d.ContinuouslyUpdateStatus()
			if err != nil {
				log.Fatalf("Error continuously updating deployment status: %s", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		err := d.ContinuouslyHandleNoDeactivationAutostart()
		if err != nil {
			log.Fatalf("Error continuously handling noDeactivation autostart: %s", err.Error())
		}
	}()
	return d, nil
}

func (d *DeploymentHandler) UpdateStatus(dpl *appsv1.Deployment) error {
	var deployment *appsv1.Deployment
	if dpl == nil {
		var err error
		deployment, err = d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).Get(context.TODO(), d.Config.Deployment, metav1.GetOptions{})
		if err != nil {
			log.Printf("Error getting deployment: %s", err.Error())
			return err
		}
	} else {
		deployment = dpl
	}
	if deployment.Status.ReadyReplicas > 0 && *deployment.Spec.Replicas > 0 {
		if d.Config.ReadinessProbePath != "" {
			targetBaseUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", d.Config.Service, d.Config.ServicePort))
			if err != nil {
				return err
			}
			readinessCheckStartTime := time.Now()
			for d.Config.ReadinessTimeoutSecs == 0 || time.Since(readinessCheckStartTime).Seconds() < float64(d.Config.ReadinessTimeoutSecs) {
				resp, err := http.Get(targetBaseUrl.String() + d.Config.ReadinessProbePath)
				if err == nil && resp.StatusCode == 200 {
					break
				}
				time.Sleep(1 * time.Second)
			}
		}
		log.Println("Deployment is ready")
		d.SetStatus(DeploymentStatusReady)
	} else if deployment.Status.Replicas > 0 && *deployment.Spec.Replicas == 0 {
		log.Println("Deployment is deactivating")
		d.SetStatus(DeploymentStatusDeactivating)
	} else if deployment.Status.Replicas == 0 && *deployment.Spec.Replicas == 0 {
		log.Println("Deployment is deactivated")
		d.SetStatus(DeploymenStatusDeactivated)
	} else if deployment.Status.ReadyReplicas == 0 && *deployment.Spec.Replicas > 0 {
		log.Println("Deployment is activating")
		d.SetStatus(DeploymentStatusActivating)
	} else {
		return errors.New("unexpected deployment status")
	}
	return nil
}

func (d *DeploymentHandler) SetStatus(status DeploymentStatus) {
	if d.Status != status {
		log.Printf("Deployment status changed from %s to %s", d.Status, status)
		d.Status = status
		d.LastStatusChange = time.Now()
	}
}

func (d *DeploymentHandler) ContinuouslyUpdateStatus() error {
	err := d.UpdateStatus(nil)
	if err != nil {
		return err
	}
	deploymentWatcher, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + d.Config.Deployment,
		Watch:         true,
	})
	if err != nil {
		log.Printf("Error creating deployment watcher: %s", err.Error())
		return err
	}
	defer deploymentWatcher.Stop()
	for event := range deploymentWatcher.ResultChan() {
		if event.Type == "MODIFIED" {
			deployment := event.Object.(*appsv1.Deployment)
			if err != nil {
				log.Printf("Error converting event object to deployment: %s", err.Error())
				return err
			}
			err := d.UpdateStatus(deployment)
			if err != nil {
				log.Printf("Error updating deployment status: %s", err.Error())
				return err
			}
		}
	}
	return nil
}

func (d *DeploymentHandler) WaitForReady() {
	for range time.Tick(250 * time.Millisecond) {
		if d.Status == DeploymentStatusReady {
			return
		}
	}
}

func (d *DeploymentHandler) ActivateDeployment() error {
	if d.Status == DeploymentStatusReady || d.Status == DeploymentStatusActivating {
		return nil
	}
	scale, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).GetScale(context.TODO(), d.Config.Deployment, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting deployment scale: %s", err.Error())
		return err
	}
	if scale.Spec.Replicas < 1 {
		scale.Spec.Replicas = 1
		_, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).UpdateScale(context.TODO(), d.Config.Deployment, scale, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("Error updating deployment scale: %s", err.Error())
			return err
		}
		d.SetStatus(DeploymentStatusActivating)
	}
	return nil
}

func (d *DeploymentHandler) DeactivateDeployment() error {
	if d.Status == DeploymenStatusDeactivated || d.Status == DeploymentStatusDeactivating {
		return nil
	}
	scale, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).GetScale(context.TODO(), d.Config.Deployment, metav1.GetOptions{})
	if err != nil {
		log.Printf("Error getting deployment scale: %s", err.Error())
		return err
	}
	if scale.Spec.Replicas > 0 {
		scale.Spec.Replicas = 0
		_, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).UpdateScale(context.TODO(), d.Config.Deployment, scale, metav1.UpdateOptions{})
		if err != nil {
			log.Printf("Error updating deployment scale: %s", err.Error())
			return err
		}
		d.SetStatus(DeploymentStatusDeactivating)
	}
	return nil
}

func (d *DeploymentHandler) ContinuouslyHandleNoDeactivationAutostart() error {
	if d.Config.NoDeactivationAutostart {
		loc, err := time.LoadLocation("UTC")
		if err != nil {
			return err
		}
		for range time.Tick(30 * time.Second) {
			if d.Status == DeploymenStatusDeactivated {
				now, err := time.ParseInLocation("15:04", time.Now().UTC().Format("15:04"), loc)
				if err != nil {
					return err
				}
				if d.Config.NoDeactivationMoFrFromToUTC != nil && (now.Weekday() == time.Monday || now.Weekday() == time.Tuesday || now.Weekday() == time.Wednesday || now.Weekday() == time.Thursday || now.Weekday() == time.Friday) {
					fromTime, err := time.ParseInLocation("15:04", d.Config.NoDeactivationMoFrFromToUTC[0], loc)
					if err != nil {
						return err
					}
					toTime, err := time.ParseInLocation("15:04", d.Config.NoDeactivationMoFrFromToUTC[1], loc)
					if err != nil {
						return err
					}
					if fromTime.Before(now) && toTime.After(now) {
						err = d.ActivateDeployment()
						if err != nil {
							return err
						}
					}
				}
				if d.Config.NoDeactivationSatFromToUTC != nil && now.Weekday() == time.Saturday {
					fromTime, err := time.ParseInLocation("15:04", d.Config.NoDeactivationSatFromToUTC[0], loc)
					if err != nil {
						return err
					}
					toTime, err := time.ParseInLocation("15:04", d.Config.NoDeactivationSatFromToUTC[1], loc)
					if err != nil {
						return err
					}
					if fromTime.Before(now) && toTime.After(now) {
						err = d.ActivateDeployment()
						if err != nil {
							return err
						}
					}
				}
				if d.Config.NoDeactivationSunFromToUTC != nil && now.Weekday() == time.Sunday {
					fromTime, err := time.ParseInLocation("15:04", d.Config.NoDeactivationSunFromToUTC[0], loc)
					if err != nil {
						return err
					}
					toTime, err := time.ParseInLocation("15:04", d.Config.NoDeactivationSunFromToUTC[1], loc)
					if err != nil {
						return err
					}
					if fromTime.Before(now) && toTime.After(now) {
						err = d.ActivateDeployment()
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
