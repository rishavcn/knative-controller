/*
Copyright 2019 Rishav Kumar.

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

package dkind

import (
	"context"
	"reflect"

	examplev1beta1 "test10/pkg/apis/example/v1beta1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1beta1 "knative.dev/serving/pkg/apis/serving/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	_ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Dkind Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDkind{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("dkind-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Dkind
	err = c.Watch(&source.Kind{Type: &examplev1beta1.Dkind{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Dkind - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &examplev1beta1.Dkind{},
	})
	if err != nil {
		return err
	}

	return nil
}

func newKnativeDeploymentForCG(ws *examplev1beta1.Dkind, labels map[string]string, message string, name string) *servingv1beta1.Service {
	// fmt.Println("#######3")
	// fmt.Println(message)
	return &servingv1beta1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1beta1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cg-sim",
			Namespace: "devflows",
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								corev1.Container{
									Image: "villardl/transformer-nodejs",
									Env: []corev1.EnvVar{{
										Name:  "TRANSFORMER",
										Value: `({"message": "processed by CG" + event.data.msg})`,
									},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// func MakeKnativeService(functionName string, version, image string) *servingv1beta1.Service {
// 	return &servingv1beta1.Service{
// 		TypeMeta: metav1.TypeMeta{
// 			APIVersion: "v1beta1",
// 			Kind:       "Service",
// 		},
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      functionName,
// 			Namespace: "knative-functions",
// 		},
// 		Spec: servingv1.ServiceSpec{
// 			ConfigurationSpec: servingv1.ConfigurationSpec{
// 				Template: servingv1.RevisionTemplateSpec{
// 					ObjectMeta: metav1.ObjectMeta{
// 						Annotations: map[string]string{duckv1alpha1.ConfigMapAnnotation: version},
// 					},
// 					Spec: servingv1.RevisionSpec{
// 						PodSpec: corev1.PodSpec{
// 							Containers: []corev1.Container{
// 								corev1.Container{
// 									Image: image,
// 									VolumeMounts: []corev1.VolumeMount{
// 										corev1.VolumeMount{
// 											Name:      "config-function-" + functionName,
// 											MountPath: "/ko-app/___config.json",
// 											SubPath:   "___config.json",
// 										},
// 									},
// 								},
// 							},
// 							Volumes: []corev1.Volume{
// 								corev1.Volume{
// 									Name: "config-function-" + functionName,
// 									VolumeSource: corev1.VolumeSource{
// 										ConfigMap: &corev1.ConfigMapVolumeSource{
// 											LocalObjectReference: corev1.LocalObjectReference{
// 												Name: "config-function-" + functionName,
// 											},
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

var _ reconcile.Reconciler = &ReconcileDkind{}

// ReconcileDkind reconciles a Dkind object
type ReconcileDkind struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileDkind) ReconcileIt(request reconcile.Request, ws *examplev1beta1.Dkind, name string) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Dkind")

	err := r.Get(context.TODO(), request.NamespacedName, ws)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	labels := map[string]string{
		"dkindlabel": ws.Name,
	}
	var message string
	if name == "2" {
		message = ws.Spec.MessageA
	} else {
		message = ws.Spec.MessageB
	}

	var deployment *servingv1beta1.Service
	// Got the Dkind resource instance, now reconcile owned Deployment and Service resources
	deployment = newKnativeDeploymentForCG(ws, labels, message, name)

	if err != nil {
		return reconcile.Result{}, err
	}

	foundDeployment := &servingv1beta1.Service{}
	// See if a Deployment already exists
	err = r.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, foundDeployment)

	// if the Deployment doesn't exist create it
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Deployment",
			"Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
		err = r.Create(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}

		//Deployment created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	//Deployment already exists, check the replica count int the status matches the desired replica count
	reqLogger.Info("Skip reconcile: Deployment already exists",
		"Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)

	// reqLogger.Info("*******")
	// fmt.Println(ws.Spec)
	// reqLogger.Info("*******")
	// fmt.Println(foundDeployment.Spec)

	if !reflect.DeepEqual(deployment.Spec, foundDeployment.Spec) {
		foundDeployment.Spec = deployment.Spec
		// log.Print("Updating Deployment %s/%s\n", deployment.Namespace, deployment.Name)
		err = r.Update(context.TODO(), foundDeployment)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

// Reconcile reads that state of the cluster for a Dkind object and makes changes based on the state read
// and what is in the Dkind.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.rishav.cn1,resources=dkinds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.rishav.cn1,resources=dkinds/status,verbs=get;update;patch
func (r *ReconcileDkind) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Dkind instance
	ws := &examplev1beta1.Dkind{}
	// fmt.Println("#######1")
	// fmt.Println(ws.Spec.MessageA)

	_, err1 := r.ReconcileIt(request, ws, "1")
	_, err2 := r.ReconcileIt(request, ws, "2")

	// fmt.Println("#######2")
	// fmt.Println(ws.Spec.MessageA)

	if err1 != nil {
		return reconcile.Result{}, err1
	}
	if err2 != nil {
		return reconcile.Result{}, err2
	}
	return reconcile.Result{}, nil
}
