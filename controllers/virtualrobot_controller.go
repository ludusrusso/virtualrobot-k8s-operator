/*


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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	robotsv1alpha1 "ludusrusso.dev/robot/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VirtualRobotReconciler reconciles a VirtualRobot object
type VirtualRobotReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=robots.ludusrusso.dev,resources=virtualrobots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=robots.ludusrusso.dev,resources=virtualrobots/status,verbs=get;update;patch

func (r *VirtualRobotReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()
	logger := r.Log.WithValues("virtualrobot", req.NamespacedName)

	var vrobot robotsv1alpha1.VirtualRobot

	if err := r.Get(ctx, req.NamespacedName, &vrobot); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Got new vrobot: %v in %v", vrobot.Name, vrobot.Namespace)

	pod, err := r.DesiredPod(vrobot)
	if err != nil {
		logger.Error(err, "Cannot create pod")
		return ctrl.Result{}, err
	}

	svc, err := r.DesiredService(vrobot)
	if err != nil {
		logger.Error(err, "Cannot create service")
		return ctrl.Result{}, err
	}

	ing, host, err := r.DesiredIngress(vrobot)
	if err != nil {
		logger.Error(err, "Cannot create ingress")
		return ctrl.Result{}, err
	}

	applyOps := []client.PatchOption{client.ForceOwnership, client.FieldOwner("virtualrobot")}
	if err := r.Patch(ctx, &pod, client.Apply, applyOps...); err != nil {
		logger.Error(err, "Cannot apply pod")
		return ctrl.Result{}, err
	}

	if err := r.Patch(ctx, &svc, client.Apply, applyOps...); err != nil {
		logger.Error(err, "Cannot apply service")
		return ctrl.Result{}, err
	}

	if err := r.Patch(ctx, &ing, client.Apply, applyOps...); err != nil {
		logger.Error(err, "Cannot apply ingress")
		return ctrl.Result{}, err
	}

	vrobot.Status.URL = host

	fmt.Printf("Host: %v in %v\n\n", vrobot.Name, vrobot.Namespace)
	err = r.Client.Status().Update(ctx, &vrobot)
	if err != nil {
		logger.Error(err, "\n\n\n\nError\n\n\n\n")

		return ctrl.Result{}, err
	}

	logger.Info("Done")

	return ctrl.Result{}, nil
}

// DesiredPod create a pod
func (r *VirtualRobotReconciler) DesiredPod(vrobot robotsv1alpha1.VirtualRobot) (corev1.Pod, error) {
	pod := corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName(vrobot),
			Namespace: vrobot.Namespace,
			Labels: map[string]string{
				"robot": string(vrobot.ObjectMeta.UID),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "bot",
					Image: "ludusrusso/vbot:v2",
					Ports: []corev1.ContainerPort{
						{
							Name:          "ws",
							ContainerPort: 9090,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(&vrobot, &pod, r.Scheme); err != nil {
		return pod, err
	}
	return pod, nil
}

// DesiredService create a Service
func (r *VirtualRobotReconciler) DesiredService(vrobot robotsv1alpha1.VirtualRobot) (corev1.Service, error) {
	svc := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildName(vrobot),
			Namespace: vrobot.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"robot": string(vrobot.ObjectMeta.UID),
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "ws",
					Port:       9090,
					TargetPort: intstr.IntOrString{IntVal: 9090},
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(&vrobot, &svc, r.Scheme); err != nil {
		return svc, err
	}
	return svc, nil
}

// DesiredIngress create an Ingress
func (r *VirtualRobotReconciler) DesiredIngress(vrobot robotsv1alpha1.VirtualRobot) (extensionsv1beta1.Ingress, string, error) {
	name := buildName(vrobot)
	host := fmt.Sprintf("%v.%v", vrobot.Name, vrobot.Spec.BaseURL)
	ing := extensionsv1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: extensionsv1beta1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: vrobot.Namespace,
			Annotations: map[string]string{
				"kubernetes.io/ingress.class":            "traefik",
				"certmanager.k8s.io/acme-challenge-type": "http01",
				"certmanager.k8s.io/cluster-issuer":      "letsencrypt-prod",
			},
		},
		Spec: extensionsv1beta1.IngressSpec{
			Rules: []extensionsv1beta1.IngressRule{
				{
					Host: host,
					IngressRuleValue: extensionsv1beta1.IngressRuleValue{
						HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
							Paths: []extensionsv1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1beta1.IngressBackend{
										ServiceName: name,
										ServicePort: intstr.IntOrString{IntVal: 9090},
									},
								},
							},
						},
					},
				},
			},
			TLS: []extensionsv1beta1.IngressTLS{
				{
					Hosts:      []string{host},
					SecretName: host,
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(&vrobot, &ing, r.Scheme); err != nil {
		return ing, "", err
	}
	return ing, host, nil
}

func (r *VirtualRobotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&robotsv1alpha1.VirtualRobot{}).
		Owns(&extensionsv1beta1.Ingress{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func buildName(vrobot robotsv1alpha1.VirtualRobot) string {
	return fmt.Sprintf("%v", vrobot.Spec.RobotName)
}
