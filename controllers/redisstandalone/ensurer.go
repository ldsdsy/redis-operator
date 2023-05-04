package redisstandalone

import (
	"context"
	redisv1 "ldsdsy/redis-operator/api/v1"

	errs "github.com/pkg/errors"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type Ensurer interface {
	ConfigMap(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error)
	// Secret(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error)
	StatefulSet(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error)
	ServiceHeadless(standalone redisv1.RedisStandalone, Labels map[string]string) (bool, error)
	ServiceNodeport(standalone redisv1.RedisStandalone, Labels map[string]string) (bool, error)
}

func (r *RedisStandaloneReconciler) EnsurerResource(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error) {
	if ok, err := r.ConfigMap(standalone, labels); !ok {
		klog.Errorln("Wrong in creating configMap of ", standalone.Name, ", err is ", err)
		return false, err
	}
	// if ok, err := r.Secret(standalone, labels); !ok {
	// 	klog.Errorln("Wrong in creating secret of ", standalone.Name, ", err is ", err)
	// 	return false, err
	// }
	if ok, err := r.StatefulSet(standalone, labels); !ok {
		klog.Errorln("Wrong in creating statefulSet of ", standalone.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.ServiceHeadless(standalone, labels); !ok {
		klog.Errorln("Wrong in creating serviceHeadless of ", standalone.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.ServiceNodeport(standalone, labels); !ok {
		klog.Errorln("Wrong in creating serviceNodeport of ", standalone.Name, ", err is ", err)
		return false, err
	}

	return true, nil
}

func (r *RedisStandaloneReconciler) ConfigMap(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error) {
	oldConfig := &corev1.ConfigMap{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: standalone.Namespace,
		Name:      standalone.Name + "-configmap",
	}, oldConfig)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newConfig, newErr := newConfigMap(standalone, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newConfigMap of ", standalone.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newConfig); err != nil {
				klog.Errorln("Wrong in creating newConfigMap of ", standalone.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr == nil {
		if newConfig, newErr := newConfigMap(standalone, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newConfigMap of ", standalone.Name, ", err is ", newErr)
			return false, newErr
		} else if isChanged(oldConfig.Data, newConfig.Data) {
			if err := r.Client.Update(context.TODO(), newConfig); err != nil {
				klog.Errorln("Wrong in updating ConfigMap of ", standalone.Name, ", err is ", err)
				return false, err
			}
		}
	} else {
		klog.Errorln("Wrong in getting oldConfigMap of ", standalone.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

// func (r *RedisStandaloneReconciler) Secret(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error) {
// 	klog.Info("creating secret!")
// 	oldSecret := &corev1.Secret{}
// 	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
// 		Namespace: standalone.Namespace,
// 		Name:      standalone.Name + "-secret",
// 	}, oldSecret)
// 	if oldErr != nil && errors.IsNotFound(oldErr) {
// 		if newSecret, newErr := newSecret(standalone, labels); newErr != nil {
// 			klog.Errorln("Wrong in Forming newSecret of ", standalone.Name, ", err is ", newErr)
// 			return false, newErr
// 		} else {
// 			if err := r.Client.Create(context.TODO(), newSecret); err != nil {
// 				klog.Errorln("Wrong in creating newSecret of ", standalone.Name, ", err is ", err)
// 				return false, err
// 			}
// 		}
// 	} else if oldErr == nil {
// 		if newSecret, newErr := newSecret(standalone, labels); newErr != nil {
// 			klog.Errorln("Wrong in Forming newSecret of ", standalone.Name, ", err is ", newErr)
// 			return false, newErr
// 		} else if isChanged(oldSecret.StringData, newSecret.StringData) {
// 			if err := r.Client.Update(context.TODO(), newSecret); err != nil {
// 				klog.Errorln("Wrong in updating Secret of ", standalone.Name, ", err is ", err)
// 				return false, err
// 			}
// 		}
// 	} else {
// 		klog.Errorln("Wrong in getting oldSecret of ", standalone.Name, ", err is ", oldErr)
// 		return false, oldErr
// 	}
// 	return true, nil
// }

func (r *RedisStandaloneReconciler) StatefulSet(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error) {
	oldStatufulSet := &appv1.StatefulSet{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: standalone.Namespace,
		Name:      standalone.Name,
	}, oldStatufulSet)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newStatufulSet, newErr := newStatufulSet(standalone, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newStatufulSet of ", standalone.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newStatufulSet); err != nil {
				klog.Errorln("Wrong in creating newStatufulSet of ", standalone.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr == nil {
		if newStatufulSet, newErr := newStatufulSet(standalone, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newStatufulSet of ", standalone.Name, ", err is ", newErr)
			return false, newErr
		} else if stsIsChanged(oldStatufulSet, newStatufulSet) {
			if err := r.Client.Update(context.TODO(), newStatufulSet); err != nil {
				klog.Errorln("Wrong in updating StatufulSet of ", standalone.Name, ", err is ", err)
				return false, err
			}
		}
	} else {
		klog.Errorln("Wrong in getting oldStatufulSet of ", standalone.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisStandaloneReconciler) ServiceHeadless(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error) {
	oldSvc := &corev1.Service{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: standalone.Namespace,
		Name:      standalone.Name,
	}, oldSvc)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newSvc, newErr := newServiceHeadless(standalone, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSvc of ", standalone.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newSvc); err != nil {
				klog.Errorln("Wrong in creating newSvc of ", standalone.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr != nil {
		klog.Errorln("Wrong in getting oldSvc of ", standalone.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisStandaloneReconciler) ServiceNodeport(standalone redisv1.RedisStandalone, labels map[string]string) (bool, error) {
	oldSvc := &corev1.Service{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: standalone.Namespace,
		Name:      standalone.Name + "-nodeport",
	}, oldSvc)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newSvc, newErr := newServiceNodeport(standalone, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSvc of ", standalone.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newSvc); err != nil {
				klog.Errorln("Wrong in creating newSvc of ", standalone.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr != nil {
		klog.Errorln("Wrong in getting oldSvc of ", standalone.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisStandaloneReconciler) CheckStatus(standalone redisv1.RedisStandalone) (bool, error) {
	redisPod := &appv1.StatefulSet{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: standalone.Namespace,
		Name:      standalone.Name,
	}, redisPod)
	if err != nil {
		klog.Errorln("Wrong in geting redisPod of ", standalone.Name, ", err is ", err)
		return false, err
	}
	// standalone mode
	if redisPod.Status.ReadyReplicas == 1 {
		return true, nil
	}
	return true, errs.New("Redis pod is not running!")
}
