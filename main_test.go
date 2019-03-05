package main

import (
	"bytes"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
)

func Test_runInjector(t *testing.T) {

	expectedOutput := `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  creationTimestamp: null
  name: test
spec:
  replicas: 1
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers:
      - image: some-image
        name: name-test
        resources: {}
      - command:
        - /cloud_sql_proxy
        - -instances=project-test:region-test:instance-test
        - -log_debug_stdout=true
        - -verbose=
        - -credential_file=/secrets/cloudsql/credentials.json
        image: gcr.io/cloudsql-docker/gce-proxy:1.11
        name: cloudsql-proxy
        resources:
          limits:
            cpu: 100m
            memory: 128Mi
          requests:
            cpu: 5m
            memory: 8Mi
        securityContext:
          allowPrivilegeEscalation: false
          runAsUser: 2
        volumeMounts:
        - mountPath: /secrets/cloudsql
          name: cloudsql-proxy-credentials
          readOnly: true
      volumes:
      - name: test-volume
        secret:
          secretName: test-secret
      - name: cloudsql-proxy-credentials
        secret:
          secretName: cloudsql-proxy-credentials
status: {}

---

apiVersion: v1
kind: Service
metadata:
  name: test-svc
spec:
  ports:
  - name: web
    port: 8080`

	// Just to trick to get control other stdout
	// r and w are linked => whatever is written in w is readable in r
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	*path = "./test/test.yaml"
	*instance = "instance-test"
	*region = "region-test"
	*project = "project-test"
	*cpuRequest = "5m"
	*memoryRequest = "8Mi"
	*cpuLimit = "100m"
	*memoryLimit = "128Mi"
	*proxyVersion = "1.11"

	runInjector()
	os.Stdout = oldStdout
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	assert.Equal(t, expectedOutput, buf.String())
}
