/*
Copyright 2022.

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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	v1 "kubevirt.io/api/core/v1"
)

// VirtualMachineInstanceReconciler reconciles a VirtualMachineInstance object
type VirtualMachineInstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// TODO ZoneManager zoneManager
}

func (r *VirtualMachineInstanceReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	vmi := &v1.VirtualMachineInstance{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, vmi)
	if err != nil {
		return ctrl.Result{}, err
	}
	// TODO zoneManager.updateVMI(vmi)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VirtualMachineInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	onVMIEvent := predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			return true
		},
		GenericFunc: func(event.GenericEvent) bool {
			return false
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.VirtualMachineInstance{}).
		WithEventFilter(onVMIEvent).
		Complete(r)
}
