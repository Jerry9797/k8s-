package main

import (
	"hook/lib"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"log"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/klog/v2"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		var body []byte
		if r.Body != nil {
			if data, err := io.ReadAll(r.Body); err == nil {
				body = data
			}
		}
		reqAdmissionReview := &v1.AdmissionReview{}
		rspAdmissionReview := &v1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
		}
		// 把 body decode 成对象
		deserializer := lib.Codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(body, nil, reqAdmissionReview); err != nil {
			klog.Error(err)
			rspAdmissionReview.Response = lib.ToV1AdmissionResponse(err)
		} else {
			rspAdmissionReview.Response = lib.AdmitPods(*reqAdmissionReview)
		}
		//返回响应
		rspBody, err := json.Marshal(rspAdmissionReview)
		if err != nil {
			log.Println("Failed to marshal response:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(rspBody)
	})

	tlsConfig := lib.Config{
		CertFile: "/etc/webhook/certs/tls.crt",
		KeyFile:  "/etc/webhook/certs/tls.key",
	}
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: lib.ConfigTLS(tlsConfig),
	}
	server.ListenAndServeTLS("", "")

	//err := http.ListenAndServe(":8080", nil)
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
}
