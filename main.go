package main

import (
	"encoding/json"
	"os"

	"github.com/go-yaml/yaml"
	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/api/apps/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	path = kingpin.Arg("path", "Deployment file path").Required().String()
)

func main() {
	kingpin.Parse()

	var cloudSQLProxyContainer v1.Container
	{
		limitCPUQt := resource.NewMilliQuantity(100, resource.BinarySI)
		limitMemoryQt := resource.NewQuantity(128*1024e3, resource.BinarySI)
		limits := v1.ResourceList{
			v1.ResourceCPU:    *limitCPUQt,
			v1.ResourceMemory: *limitMemoryQt,
		}
		requestCPUQt := resource.NewMilliQuantity(5, resource.BinarySI)
		requestMemoryQt := resource.NewMilliQuantity(8, resource.BinarySI)
		requests := v1.ResourceList{
			v1.ResourceCPU:    *requestCPUQt,
			v1.ResourceMemory: *requestMemoryQt,
		}

		var runAsUser int64 = 2
		var allowPrivilegeEscalation = false

		securityContext := v1.SecurityContext{
			RunAsUser:                &runAsUser,
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		}

		volumeMount := v1.VolumeMount{
			Name:      "cloudsql-proxy-credentials",
			MountPath: "/secrets/cloudsql",
			ReadOnly:  true,
		}

		cloudSQLProxyContainer = v1.Container{}
		cloudSQLProxyContainer.Name = "cloudsql-proxy"
		cloudSQLProxyContainer.Image = "gcr.io/cloudsql-docker/gce-proxy:1.11"
		cloudSQLProxyContainer.Command = []string{"/cloud_sql_proxy", "-instances=ricardo-dev-ch:europe-west1:ricardo-dev-postgres=tcp:5432", "-credential_file=/secrets/cloudsql/credentials.json"}
		cloudSQLProxyContainer.Resources = v1.ResourceRequirements{Limits: limits, Requests: requests}
		cloudSQLProxyContainer.SecurityContext = &securityContext
		cloudSQLProxyContainer.VolumeMounts = []v1.VolumeMount{volumeMount}
	}

	b, err := yaml.Marshal(&cloudSQLProxyContainer)
	if err != nil {
		panic(err)
	}
	f, err := os.Open(*path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, _ = json.Marshal(&cloudSQLProxyContainer)

	deploy := &v1beta1.Deployment{}
	err = json.Unmarshal(b, &deploy)
	if err != nil {
		panic(err)
	}
	k8syaml.NewYAMLOrJSONDecoder(f, 4096).Decode(&deploy)

	deploy.Spec.Template.Spec.Volumes = []v1.Volume{v1.Volume{
		Name: "cloudsql-proxy-credentials",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: "cloudsql-proxy-credentials",
			},
		},
	}}
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, cloudSQLProxyContainer)

	serializer := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, nil, nil)
	serializer.Encode(deploy, os.Stdout)
}
