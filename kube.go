package main

import (
	"log"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	rules      = clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
)

type Versions struct {
	appName      string
	appNS        string
	operatorName string
	operatorNS   string
	clientset    kubernetes.Clientset
}

// Initalizes the kube environment, checks for a kube config, if one doesn't
// exist an error is printed to the screen.
func initKube() {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			handlePanicTop(err)
		}
	}()

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	v := Versions{
		appName:      "nginx-deployment",
		appNS:        "default",
		operatorName: "coredns",
		operatorNS:   "kube-system",
		clientset:    *clientset,
	}
	v.getVersions()

}

func (v *Versions) getVersions() {

	appDeploy, err := v.clientset.AppsV1().Deployments(v.appNS).Get(ctx, v.appName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
	}
	appImage := appDeploy.Spec.Template.Spec.Containers[0].Image
	appVersion := strings.Split(appImage, ":")

	operatorDeploy, err := v.clientset.AppsV1().Deployments(v.operatorNS).Get(ctx, v.operatorName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
	}
	operatorImage := operatorDeploy.Spec.Template.Spec.Containers[0].Image
	operatorVersion := strings.Split(operatorImage, ":")

	final := []string{"cnvrg-app version: " + appVersion[len(appVersion)-1], "operator version: " + operatorVersion[len(operatorVersion)-1]}

	a := strings.Join(final, "\n")
	setTopText(a, "white")

}

func pgBackup() {
	InfoLogger.Println("In the pg backup func")
}
