package redissentinel

import (
	"bytes"
	"encoding/json"
	redisv1 "ldsdsy/redis-operator/api/v1"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func DefaultOwnerReferences(sentinel redisv1.RedisSentinel) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(&sentinel, schema.GroupVersionKind{
			Group:   "redis.my.domain",
			Version: "v1",
			Kind:    "RedisSentinel",
		}),
	}
}

func newRedisConfigMap(sentinel redisv1.RedisSentinel, labels map[string]string) (*corev1.ConfigMap, error) {
	parameters := transform(sentinel.Spec.Redis.Configuration)
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name + "-redis-configmap",
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Data: map[string]string{
			"redis.conf": parameters,
		},
	}
	return config, nil
}
func newSentinelConfigMap(sentinel redisv1.RedisSentinel, labels map[string]string) (*corev1.ConfigMap, error) {
	parameters := transform(sentinel.Spec.Sentinel.Configuration)
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name + "-sentinel-configmap",
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Data: map[string]string{
			"sentinel.conf": parameters,
		},
	}
	return config, nil
}

func transform(p map[string]string) string {
	var parameters string
	for k, v := range p {
		parameters = parameters + k + " " + v + "\n"

	}
	return parameters
}

func isChanged(old map[string]string, new map[string]string) bool {
	if len(old) != len(new) {
		return true
	}
	for k, v1 := range old {
		if v2, ok := new[k]; !ok {
			return true
		} else {
			if v1 != v2 {
				return true
			}
		}
	}
	return false
}

func stsIsChanged(old *appv1.StatefulSet, new *appv1.StatefulSet) bool {
	if old.Spec.Replicas != new.Spec.Replicas {
		return true
	}
	//replicas', 'template', 'updateStrategy', 'persistentVolumeClaimRetentionPolicy' and 'minReadySeconds'
	oldTemplate, _ := json.Marshal(old.Spec.Template)
	newTemplate, _ := json.Marshal(new.Spec.Template)
	if !bytes.Equal(oldTemplate, newTemplate) {
		return true
	}
	oldPolicy, _ := json.Marshal(old.Spec.PersistentVolumeClaimRetentionPolicy)
	newPolicy, _ := json.Marshal(new.Spec.PersistentVolumeClaimRetentionPolicy)
	return !bytes.Equal(oldPolicy, newPolicy)
}

func newRedisStatufulSet(sentinel redisv1.RedisSentinel, labels map[string]string) (*appv1.StatefulSet, error) {
	labels["identity"] = "redis"
	statufulSet := &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name,
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Spec: appv1.StatefulSetSpec{
			ServiceName: sentinel.Name,
			Replicas:    &sentinel.Spec.Redis.Replicas,
			PersistentVolumeClaimRetentionPolicy: &appv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: "Delete",
				WhenScaled:  "Delete",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						rediscontainer(sentinel),
					},
					Volumes: []corev1.Volume{
						{
							Name: "redis-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: sentinel.Name + "-redis-configmap",
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:            "redis-data",
						OwnerReferences: DefaultOwnerReferences(sentinel),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &sentinel.Spec.Redis.Storage.StorageClass,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: sentinel.Spec.Redis.Storage.Size,
							},
						},
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
					},
				},
			},
		},
	}
	return statufulSet, nil
}

func rediscontainer(sentinel redisv1.RedisSentinel) corev1.Container {
	container := corev1.Container{
		Name:            "redis",
		Image:           sentinel.Spec.Redis.Image,
		ImagePullPolicy: sentinel.Spec.Redis.ImagePullPolicy,
		Command: []string{
			"redis-server",
			"/conf/redis.conf",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "client",
				ContainerPort: 6379,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "redis-data",
				MountPath: "/data",
			},
			{
				Name:      "redis-conf",
				MountPath: "/conf",
			},
		},
		Resources: sentinel.Spec.Redis.Resources,
	}
	return container
}
func newSentinelStatufulSet(sentinel redisv1.RedisSentinel, labels map[string]string) (*appv1.StatefulSet, error) {
	labels["identity"] = "sentinel"
	statufulSet := &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name + "-sen",
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Spec: appv1.StatefulSetSpec{
			ServiceName: sentinel.Name,
			Replicas:    &sentinel.Spec.Sentinel.Replicas,
			PersistentVolumeClaimRetentionPolicy: &appv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: "Delete",
				WhenScaled:  "Delete",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						sentinelcontainer(sentinel),
					},
					Volumes: []corev1.Volume{
						{
							Name: "sentinel-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: sentinel.Name + "-sentinel-configmap",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return statufulSet, nil
}
func sentinelcontainer(sentinel redisv1.RedisSentinel) corev1.Container {
	container := corev1.Container{
		Name:            "sentinel",
		Image:           sentinel.Spec.Sentinel.Image,
		ImagePullPolicy: sentinel.Spec.Sentinel.ImagePullPolicy,
		Command: []string{
			"redis-server",
			"/conf/sentinel.conf",
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          "client",
				ContainerPort: 26379,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "sentinel-conf",
				MountPath: "/conf",
			},
		},
		Resources: sentinel.Spec.Redis.Resources,
	}
	return container
}

func newRedisServiceHeadless(sentinel redisv1.RedisSentinel, labels map[string]string) (*corev1.Service, error) {
	labels["identity"] = "redis"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name,
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "redis",
					Port: 6379,
				},
			},
			ClusterIP: corev1.ClusterIPNone,
		},
	}
	return service, nil
}
func newSentinelServiceHeadless(sentinel redisv1.RedisSentinel, labels map[string]string) (*corev1.Service, error) {
	labels["identity"] = "sentinel"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name + "-sen",
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "sentinel",
					Port: 26379,
				},
			},
			ClusterIP: corev1.ClusterIPNone,
		},
	}
	return service, nil
}

func newServiceNodeport(sentinel redisv1.RedisSentinel, labels map[string]string) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sentinel.Name + "-master",
			Namespace:       sentinel.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(sentinel),
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name: "redis",
					Port: 6379,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	return service, nil
}

func needUpadteCR(old, new *redisv1.RedisSentinelStatus) bool {
	klog.Infoln("old is ", old, " new is ", new)
	if old.Reason != new.Reason {
		klog.Info("test1")
		return true
	}
	oldStatus, _ := json.Marshal(old.Status)
	newStatus, _ := json.Marshal(new.Status)
	return !bytes.Equal(oldStatus, newStatus)
}
