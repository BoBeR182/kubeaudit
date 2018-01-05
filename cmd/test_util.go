package cmd

import (
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
)

var path = "../fixtures/"

func runAuditTest(t *testing.T, file string, function interface{}, errCodes []int, argStr ...string) (results []Result) {
	assert := assert.New(t)
	file = filepath.Join(path, file)
	var image imgFlags
	var limits limitFlags
	switch function.(type) {
	case (func(imgFlags, k8sRuntime.Object) []Result):
		if len(argStr) != 1 {
			log.Fatal("Incorrect number of images specified")
		}
		image = imgFlags{img: argStr[0]}
		image.splitImageString()
	case (func(limitFlags, k8sRuntime.Object) []Result):
		if len(argStr) == 2 {
			limits = limitFlags{cpuArg: argStr[0], memoryArg: argStr[1]}
		} else if len(argStr) == 0 {
			limits = limitFlags{cpuArg: "", memoryArg: ""}
		} else {
			log.Fatal("Incorrect number of images specified")
		}
		limits.parseLimitFlags()
	}

	resources, err := getKubeResourcesManifest(file)
	assert.Nil(err)

	for _, resource := range resources {
		var currentResults []Result
		switch f := function.(type) {
		case (func(k8sRuntime.Object) []Result):
			currentResults = f(resource)
		case (func(imgFlags, k8sRuntime.Object) []Result):
			currentResults = f(image, resource)
		case (func(limitFlags, k8sRuntime.Object) []Result):
			currentResults = f(limits, resource)
		default:
			log.Fatal("Invalid function provided")
		}
		for _, currentResult := range currentResults {
			results = append(results, currentResult)
		}
	}
	errors := map[int]bool{}
	for _, result := range results {
		for _, occurrence := range result.Occurrences {
			errors[occurrence.id] = true
		}
	}

	for _, errCode := range errCodes {
		assert.True(errors[errCode])
	}
	assert.Equal(len(errors), len(errCodes))
	return
}

func runAuditTestInNamespace(t *testing.T, namespace string, file string, function interface{}, errCodes []int) {
	rootConfig.namespace = namespace
	runAuditTest(t, file, function, errCodes)
	rootConfig.namespace = apiv1.NamespaceAll
}
