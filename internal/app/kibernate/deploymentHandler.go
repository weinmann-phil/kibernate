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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}
	d := &DeploymentHandler{Config: config, KubeClientSet: clientSet}
	err = d.UpdateStatus(nil)
	if err != nil {
		return nil, err
	}
	go func() {
		err := d.ContinuouslyUpdateStatus()
		if err != nil {
			panic(err)
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
			return err
		}
	} else {
		deployment = dpl
	}
	if deployment.Status.ReadyReplicas > 0 && *deployment.Spec.Replicas > 0 {
		d.SetStatus(DeploymentStatusReady)
	} else if deployment.Status.Replicas > 0 && *deployment.Spec.Replicas == 0 {
		d.SetStatus(DeploymentStatusDeactivating)
	} else if deployment.Status.Replicas == 0 && *deployment.Spec.Replicas == 0 {
		d.SetStatus(DeploymenStatusDeactivated)
	} else if deployment.Status.ReadyReplicas == 0 && *deployment.Spec.Replicas > 0 {
		d.SetStatus(DeploymentStatusActivating)
	} else {
		return errors.New("unexpected deployment status")
	}
	return nil
}

func (d *DeploymentHandler) SetStatus(status DeploymentStatus) {
	if d.Status != status {
		d.Status = status
		d.LastStatusChange = time.Now()
	}
}

func (d *DeploymentHandler) ContinuouslyUpdateStatus() error {
	deploymentWatcher, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: "metadata.name=" + d.Config.Deployment,
		Watch:         true,
	})
	if err != nil {
		return err
	}
	defer deploymentWatcher.Stop()
	for event := range deploymentWatcher.ResultChan() {
		if event.Type == "MODIFIED" {
			deployment := event.Object.(*appsv1.Deployment)
			if err != nil {
				return err
			}
			err := d.UpdateStatus(deployment)
			if err != nil {
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
		return err
	}
	if scale.Spec.Replicas < 1 {
		scale.Spec.Replicas = 1
		_, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).UpdateScale(context.TODO(), d.Config.Deployment, scale, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DeploymentHandler) DeactivateDeployment() error {
	if d.Status == DeploymenStatusDeactivated || d.Status == DeploymentStatusDeactivating {
		return nil
	}
	scale, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).GetScale(context.TODO(), d.Config.Deployment, metav1.GetOptions{})
	if err != nil {
		return err
	}
	if scale.Spec.Replicas > 0 {
		scale.Spec.Replicas = 0
		_, err := d.KubeClientSet.AppsV1().Deployments(d.Config.Namespace).UpdateScale(context.TODO(), d.Config.Deployment, scale, metav1.UpdateOptions{})
		return err
	}
	return nil
}
