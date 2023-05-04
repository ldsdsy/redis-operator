/*
Copyright 2022 ldsdsy.

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

package redissentinel

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	redisv1 "ldsdsy/redis-operator/api/v1"
)

// RedisSentinelReconciler reconciles a RedisSentinel object
type RedisSentinelReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=redis.my.domain,resources=redissentinels,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redis.my.domain,resources=redissentinels/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redis.my.domain,resources=redissentinels/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisSentinel object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *RedisSentinelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	instance := &redisv1.RedisSentinel{}
	if err := r.Client.Get(context.TODO(), req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			klog.Errorln("Not found RedisSentinel: ", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		klog.Errorln("Wrong in getting RedisSentinel: ", req.NamespacedName)
		return ctrl.Result{}, err
	}
	labels := map[string]string{
		"instance": instance.Namespace + "_" + instance.Name,
	}
	// config,pvc,deploy,svc
	if ok, err := r.EnsurerResource(*instance, labels); !ok {
		klog.Errorln("Wrong in creating resources of ", instance.Name, ", err is ", err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}

	// checkout status of redis
	if ok, err := r.CheckStatus(*instance); !ok {
		klog.Errorln("Wrong in checking status of ", instance.Name, ", err is ", err)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	} else {
		newStatus := instance.Status.DeepCopy()
		if err != nil {
			newStatus.Status = redisv1.StatusKO
			newStatus.Reason = err.Error()
		} else {
			newStatus.Status = redisv1.StatusOK
			newStatus.Reason = "OK"
		}
		//加一个判断看是否需要更新
		if needUpadteCR(&instance.Status, newStatus) {
			instance.Status = *newStatus
			if err := r.Status().Update(context.TODO(), instance); err != nil {
				klog.Errorln("Wrong in updating status of ", instance.Name, ", err is ", err)
			}
		}

	}
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisSentinelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisv1.RedisSentinel{}).
		Complete(r)
}
