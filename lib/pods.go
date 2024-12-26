package lib

import (
	"fmt"
	"k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strings"
)

// only allow pods to pull images from specific registry.
func AdmitPods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("admitting pods")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		err := fmt.Errorf("expect resource to be %s", podResource)
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	fmt.Println(pod)
	reviewResponse := v1.AdmissionResponse{
		UID: ar.Request.UID, // 确保将请求中的 UID 复制到响应中
	}

	// 当pod.Name == "heyilu" 时，禁止创建pod
	if pod.Name == "heyilu" {
		reviewResponse.Allowed = false
		reviewResponse.Result = &metav1.Status{Status: "Failure", Message: strings.TrimSpace("pod name cannot be heyilu"), Code: 503}
	} else {
		reviewResponse.Allowed = true
		// 修改镜像
		reviewResponse.Patch = patchImage()
		json := v1.PatchTypeJSONPatch
		reviewResponse.PatchType = &json
	}
	return &reviewResponse
}

func patchImage() []byte {
	/*
		"op": "replace" 代替原有镜像；"op": "add" 增加镜像
	*/
	str := `[
		{
			"op": "replace",
			"path": "/spec/containers/0/image",
			"value": "nginx:1.19-alpine"
		},
		{
			"op": "add",
			"path": "/spec/initContainers",
			"value": [
				"name": "myinit",
				"image": "busybox:1.28",
				"command": ["sh","-c","echo The app is running"]
			]
		}
	]`
	return []byte(str)
}
