package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

var (
	path          = kingpin.Flag("path", "Deployment file path where to inject clousql proxy (eg. ./my-deploy-manifest.yaml)").Required().String()
	instance      = kingpin.Flag("instance", "CloudSQL instance (eg. my-clousql-instance=tcp:5432)").Required().String()
	region        = kingpin.Flag("region", "GCP region (eg. europe-west1)").Required().String()
	project       = kingpin.Flag("project", "GCP project ID (eg. ricardo)").Required().String()
	cpuRequest    = kingpin.Flag("cpu-request", "CPU request of the sidecar container").Default("5m").String()
	memoryRequest = kingpin.Flag("memory-request", "Memory request of the sidecar container").Default("8Mi").String()
	cpuLimit      = kingpin.Flag("cpu-limit", "CPU limit of the sidecar container").Default("100m").String()
	memoryLimit   = kingpin.Flag("memory-limit", "Memory limit of the sidecar container").Default("128Mi").String()
	proxyVersion  = kingpin.Flag("proxy-version", "CloudSQL proxy version").Default("1.11").String()
	verbose       = kingpin.Flag("verbose", "CloudSQL proxy verbose mode").Default("false").String()
)

func main() {
	kingpin.Parse()
	runInjector()
}

func runInjector() {
	cloudSQLProxyContainer := getCloudContainer()

	// split the file bytes by resources
	// a file may contains multiple resources, separated by "---"
	allK8SResources := getAllResourcesBytes(*path)
	// separate deployment from others resources
	deploymentBytes, otherResources := extractDeploymentBytes(allK8SResources)

	deploy := v1beta1.Deployment{}
	err := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(deploymentBytes), 4096).Decode(&deploy)
	if err != nil {
		panic(err)
	}

	deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, v1.Volume{
		Name: "cloudsql-proxy-credentials",
		VolumeSource: v1.VolumeSource{
			Secret: &v1.SecretVolumeSource{
				SecretName: "cloudsql-proxy-credentials",
			},
		},
	})
	deploy.Spec.Template.Spec.Containers = append(deploy.Spec.Template.Spec.Containers, cloudSQLProxyContainer)

	serializer := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, nil, nil)

	outputBytes := bytes.NewBuffer(nil)
	serializer.Encode(&deploy, outputBytes)
	putItBack(otherResources, outputBytes)

	os.Stdout.Write(outputBytes.Bytes())
}

func setResources(cpuRequest, memoryRequest, cpuLimit, memoryLimit string) (request v1.ResourceList, limit v1.ResourceList) {
	requestCPU, err := resource.ParseQuantity(cpuRequest)
	if err != nil {
		panic(err)
	}
	requestMemory, err := resource.ParseQuantity(memoryRequest)
	if err != nil {
		panic(err)
	}
	request = v1.ResourceList{
		v1.ResourceCPU:    requestCPU,
		v1.ResourceMemory: requestMemory,
	}

	limitCPU, err := resource.ParseQuantity(cpuLimit)
	if err != nil {
		panic(err)
	}
	limitMemory, err := resource.ParseQuantity(memoryLimit)
	if err != nil {
		panic(err)
	}

	limit = v1.ResourceList{
		v1.ResourceCPU:    limitCPU,
		v1.ResourceMemory: limitMemory,
	}

	return request, limit
}

func getAllResourcesBytes(filepath string) [][]byte {
	f, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fileBytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return bytes.Split(fileBytes, []byte("\n---"))
}

func extractDeploymentBytes(allK8SResources [][]byte) (deploymentBytes []byte, otherResources [][]byte) {
	// find the deployment in the list of resources
	for _, resourceBytes := range allK8SResources {
		// Because interpreter read only JSON...
		resourceJSON, err := k8syaml.ToJSON(resourceBytes)
		if err != nil {
			panic(err)
		}
		schema, err := k8sjson.DefaultMetaFactory.Interpret(resourceJSON)
		if err != nil {
			panic(err)
		}

		// Is this a deployment or something else
		if schema.Kind == "Deployment" {
			deploymentBytes = resourceBytes
		} else {
			otherResources = append(otherResources, resourceBytes)
		}
	}

	if len(deploymentBytes) <= 0 {
		panic(errors.New("could not find deployment resource in given file"))
	}

	return deploymentBytes, otherResources
}

func getCloudContainer() v1.Container {
	var cloudSQLProxyContainer v1.Container
	{
		requestResources, limitResources := setResources(*cpuRequest, *memoryRequest, *cpuLimit, *memoryLimit)

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
		cloudSQLProxyContainer.Image = fmt.Sprintf("gcr.io/cloudsql-docker/gce-proxy:%s", *proxyVersion)
		cloudSQLProxyContainer.Command = []string{"/cloud_sql_proxy", fmt.Sprintf("-instances=%s:%s:%s", *project, *region, *instance), "-log_debug_stdout=true", fmt.Sprintf("-verbose=%s", *verbose), "-credential_file=/secrets/cloudsql/credentials.json"}
		cloudSQLProxyContainer.Resources = v1.ResourceRequirements{Requests: requestResources, Limits: limitResources}
		cloudSQLProxyContainer.SecurityContext = &securityContext
		cloudSQLProxyContainer.VolumeMounts = append(cloudSQLProxyContainer.VolumeMounts, volumeMount)
	}


	return cloudSQLProxyContainer
}

// Put the remaining bytes that are not the deployment, back in the output
func putItBack(otherResources [][]byte, w io.Writer) {
	for _, resourceBytes := range otherResources {
		w.Write([]byte("\n---\n"))
		w.Write(resourceBytes)
	}
}
