package redisstandalone

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

func DefaultOwnerReferences(standalone redisv1.RedisStandalone) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(&standalone, schema.GroupVersionKind{
			Group:   "redis.my.domain",
			Version: "v1",
			Kind:    "RedisStandalone",
		}),
	}
}

func newConfigMap(standalone redisv1.RedisStandalone, labels map[string]string) (*corev1.ConfigMap, error) {
	parameters := transform(standalone.Spec.Configuration)
	config := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            standalone.Name + "-configmap",
			Namespace:       standalone.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(standalone),
		},
		Data: map[string]string{
			"redis.conf": parameters,
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

// func newSecret(standalone redisv1.RedisStandalone, labels map[string]string) (*corev1.Secret, error) {
// 	secret := &corev1.Secret{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:            standalone.Name + "-secret",
// 			Namespace:       standalone.Namespace,
// 			Labels:          labels,
// 			OwnerReferences: DefaultOwnerReferences(standalone),
// 		},
// 		StringData: map[string]string{
// 			"password": standalone.Spec.Password,
// 		},
// 	}
// 	return secret, nil
// }

func newStatufulSet(standalone redisv1.RedisStandalone, labels map[string]string) (*appv1.StatefulSet, error) {
	var replicas int32 = 1
	statufulSet := &appv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:            standalone.Name,
			Namespace:       standalone.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(standalone),
		},
		Spec: appv1.StatefulSetSpec{
			ServiceName: standalone.Name,
			Replicas:    &replicas,
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
						container(standalone),
					},
					Volumes: []corev1.Volume{
						{
							Name: "redis-conf",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: standalone.Name + "-configmap",
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
						OwnerReferences: DefaultOwnerReferences(standalone),
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						StorageClassName: &standalone.Spec.Storage.StorageClass,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: standalone.Spec.Storage.Size,
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

func container(standalone redisv1.RedisStandalone) corev1.Container {
	container := corev1.Container{
		Name:            "redis",
		Image:           standalone.Spec.Image,
		ImagePullPolicy: standalone.Spec.ImagePullPolicy,
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
		Resources: standalone.Spec.Resources,
	}
	return container
}

func newServiceHeadless(standalone redisv1.RedisStandalone, labels map[string]string) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            standalone.Name,
			Namespace:       standalone.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(standalone),
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

func newServiceNodeport(standalone redisv1.RedisStandalone, labels map[string]string) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            standalone.Name + "-nodeport",
			Namespace:       standalone.Namespace,
			Labels:          labels,
			OwnerReferences: DefaultOwnerReferences(standalone),
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

func needUpadteCR(old, new *redisv1.RedisStandaloneStatus) bool {
	klog.Infoln("old is ", old, " new is ", new)
	if old.Reason != new.Reason {
		klog.Info("test1")
		return true
	}
	oldStatus, _ := json.Marshal(old.Status)
	newStatus, _ := json.Marshal(new.Status)
	return !bytes.Equal(oldStatus, newStatus)
}
