/*
Copyright 2019 Pressinfra SRL.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wordpress

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logf "github.com/presslabs/controller-util/log"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

func TestPodTemplate(t *testing.T) {
	klog.SetOutput(GinkgoWriter)
	logf.SetLogger(klogr.New())

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Pod Template Suite", []Reporter{printer.NewlineReporter{}})
}
