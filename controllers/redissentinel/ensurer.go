package redissentinel

import (
	"context"
	redisv1 "ldsdsy/redis-operator/api/v1"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type Ensurer interface {
	RedisConfigMap(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error)
	SentinelConfigMap(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error)
	RedisStatefulSet(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error)
	SentinelStatefulSet(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error)
	RedisServiceHeadless(sentinel redisv1.RedisSentinel, Labels map[string]string) (bool, error)
	SentinelServiceHeadless(sentinel redisv1.RedisSentinel, Labels map[string]string) (bool, error)
	MasterServiceNodeport(sentinel redisv1.RedisSentinel, Labels map[string]string) (bool, error)
}

func (r *RedisSentinelReconciler) EnsurerResource(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	if ok, err := r.RedisConfigMap(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating redisConfigMap of ", sentinel.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.SentinelConfigMap(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating sentinelConfigMap of ", sentinel.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.RedisStatefulSet(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating redisStatefulSet of ", sentinel.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.SentinelStatefulSet(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating sentinelStatefulSet of ", sentinel.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.RedisServiceHeadless(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating redisServiceHeadless of ", sentinel.Name, ", err is ", err)
		return false, err
	}
	if ok, err := r.SentinelServiceHeadless(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating sentinelServiceHeadless of ", sentinel.Name, ", err is ", err)
		return false, err
	}

	if ok, err := r.MasterServiceNodeport(sentinel, labels); !ok {
		klog.Errorln("Wrong in creating masterServiceNodeport of ", sentinel.Name, ", err is ", err)
		return false, err
	}

	return true, nil
}

func (r *RedisSentinelReconciler) RedisConfigMap(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldConfig := &corev1.ConfigMap{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name + "-redis-configmap",
	}, oldConfig)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newConfig, newErr := newRedisConfigMap(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newRedisConfigMap of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newConfig); err != nil {
				klog.Errorln("Wrong in creating newRedisConfigMap of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr == nil {
		if newConfig, newErr := newRedisConfigMap(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newRedisConfigMap of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else if isChanged(oldConfig.Data, newConfig.Data) {
			if err := r.Client.Update(context.TODO(), newConfig); err != nil {
				klog.Errorln("Wrong in updating redisConfigMap of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else {
		klog.Errorln("Wrong in getting oldRedisConfigMap of ", sentinel.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisSentinelReconciler) SentinelConfigMap(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldConfig := &corev1.ConfigMap{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name + "-sentinel-configmap",
	}, oldConfig)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newConfig, newErr := newSentinelConfigMap(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSentinelConfigMap of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newConfig); err != nil {
				klog.Errorln("Wrong in creating newSentinelConfigMap of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr == nil {
		if newConfig, newErr := newSentinelConfigMap(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSentinelConfigMap of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else if isChanged(oldConfig.Data, newConfig.Data) {
			if err := r.Client.Update(context.TODO(), newConfig); err != nil {
				klog.Errorln("Wrong in updating sentinelConfigMap of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else {
		klog.Errorln("Wrong in getting oldSentinelConfigMap of ", sentinel.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisSentinelReconciler) RedisStatefulSet(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldStatufulSet := &appv1.StatefulSet{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name,
	}, oldStatufulSet)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newStatufulSet, newErr := newRedisStatufulSet(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newStatufulSet of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newStatufulSet); err != nil {
				klog.Errorln("Wrong in creating newStatufulSet of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr == nil {
		if newStatufulSet, newErr := newRedisStatufulSet(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newStatufulSet of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else if stsIsChanged(oldStatufulSet, newStatufulSet) {
			if err := r.Client.Update(context.TODO(), newStatufulSet); err != nil {
				klog.Errorln("Wrong in updating StatufulSet of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else {
		klog.Errorln("Wrong in getting oldStatufulSet of ", sentinel.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}
func (r *RedisSentinelReconciler) SentinelStatefulSet(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldStatufulSet := &appv1.StatefulSet{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name + "-sen",
	}, oldStatufulSet)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newStatufulSet, newErr := newSentinelStatufulSet(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSentinelStatufulSet of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newStatufulSet); err != nil {
				klog.Errorln("Wrong in creating newSentinelStatufulSet of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr == nil {
		if newStatufulSet, newErr := newSentinelStatufulSet(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSentinelStatufulSet of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else if stsIsChanged(oldStatufulSet, newStatufulSet) {
			if err := r.Client.Update(context.TODO(), newStatufulSet); err != nil {
				klog.Errorln("Wrong in updating sentinelStatufulSet of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else {
		klog.Errorln("Wrong in getting oldSentinelStatufulSet of ", sentinel.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisSentinelReconciler) RedisServiceHeadless(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldSvc := &corev1.Service{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name,
	}, oldSvc)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newSvc, newErr := newRedisServiceHeadless(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSvc of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newSvc); err != nil {
				klog.Errorln("Wrong in creating newRedisSvc of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr != nil {
		klog.Errorln("Wrong in getting oldSvc of ", sentinel.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisSentinelReconciler) SentinelServiceHeadless(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldSvc := &corev1.Service{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name + "-sen",
	}, oldSvc)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newSvc, newErr := newSentinelServiceHeadless(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newSvc of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newSvc); err != nil {
				klog.Errorln("Wrong in creating newSentinelSvc of ", sentinel.Name+"-sen", ", err is ", err)
				return false, err
			}
		}
	} else if oldErr != nil {
		klog.Errorln("Wrong in getting oldSvc of ", sentinel.Name+"-sen", ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisSentinelReconciler) MasterServiceNodeport(sentinel redisv1.RedisSentinel, labels map[string]string) (bool, error) {
	oldSvc := &corev1.Service{}
	oldErr := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: sentinel.Namespace,
		Name:      sentinel.Name + "-master",
	}, oldSvc)
	if oldErr != nil && errors.IsNotFound(oldErr) {
		if newSvc, newErr := newServiceNodeport(sentinel, labels); newErr != nil {
			klog.Errorln("Wrong in Forming newMasterSvc of ", sentinel.Name, ", err is ", newErr)
			return false, newErr
		} else {
			if err := r.Client.Create(context.TODO(), newSvc); err != nil {
				klog.Errorln("Wrong in creating newMasterSvc of ", sentinel.Name, ", err is ", err)
				return false, err
			}
		}
	} else if oldErr != nil {
		klog.Errorln("Wrong in getting oldMasterSvc of ", sentinel.Name, ", err is ", oldErr)
		return false, oldErr
	}
	return true, nil
}

func (r *RedisSentinelReconciler) CheckStatus(sentinel redisv1.RedisSentinel) (bool, error) {
	//logic
	return true, nil
}
