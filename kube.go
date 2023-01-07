package main

import (
	"fmt"
	"log"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	rules      = clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, _  = kubeconfig.ClientConfig()
	clientset  = kubernetes.NewForConfigOrDie(config)
)

type Versions struct {
	appName      string
	appNS        string
	operatorName string
	operatorNS   string
	kube         string
}

// TODO Add function that checks if a kubeconfig file exisits and don't panic, just print error to screen

func initKube() {

	v := Versions{
		appName:      "nginx-deployment",
		appNS:        "default",
		operatorName: "coredns",
		operatorNS:   "kube-system",
	}
	v.getVersions()

}

func (v *Versions) getVersions() {

	appDeploy, err := clientset.AppsV1().Deployments(v.appNS).Get(ctx, v.appName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
	}
	appImage := appDeploy.Spec.Template.Spec.Containers[0].Image
	appVersion := strings.Split(appImage, ":")

	operatorDeploy, err := clientset.AppsV1().Deployments(v.operatorNS).Get(ctx, v.operatorName, metav1.GetOptions{})
	if err != nil {
		log.Println(err)
	}
	operatorImage := operatorDeploy.Spec.Template.Spec.Containers[0].Image
	operatorVersion := strings.Split(operatorImage, ":")

	someText := "Here is some text"

	final := []string{someText, "cnvrg-app version: " + appVersion[len(appVersion)-1], "operator version: " + operatorVersion[len(operatorVersion)-1]}

	a := strings.Join(final, "\n")
	setText(a, "white")

	/*
		app, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, p := range app.Items {
			setText("cnvrg-app version:"+p.Spec.Containers[0].Image, "white")
			log.Println(p.Spec.Containers[0].Image)
		}
	*/

}

func getNodes() {
	nodeList, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, n := range nodeList.Items {
		fmt.Println(n.Name)
	}
}

func createPod() {
	newPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "busybox", Image: "busybox:latest", Command: []string{"sleep", "1000000"}},
			},
		},
	}

	if checkPodExists("test-pod", "default") {
		fmt.Println("Pod already exists")
	}

	if !checkPodExists("test-pod", "default") {
		pod, err := clientset.CoreV1().Pods("default").Create(ctx, newPod, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}
		fmt.Printf("Pod create, %v", pod.Name)
	}
}

func checkPodExists(name string, namespace string) bool {

	result := false

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range pods.Items {
		if p.Name == name {
			fmt.Printf("The pod already exists named: %v\n", p.Name)
			result = true
		}
	}
	return result
}

// Pass in the name of the pod, Namespace of the Pod
func getImage(name string, namespace string) {

	pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range pods.Items {
		setText("cnvrg app version:"+p.Spec.Containers[0].Image, "white")
		log.Println(p.Spec.Containers[0].Image)
	}
}

/*
func getResourcesDynamically(dynamic dynamic.Interface, ctx context.Context,
	group string, version string, resource string, namespace string) (
	[]unstructured.Unstructured, error) {

	resourceId := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	list, err := dynamic.Resource(resourceId).Namespace(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func GetDeployments(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) ([]v1.Deployment, error) {

	list, err := clientset.AppsV1().Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

/*
func getPvc(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) {

	//var ns, label, field string

	ns := "cnvrg"
	label := "cnvrg-control-plane"
	field := ""

	flag.StringVar(&ns, "namespace", "", "namespace")
	flag.StringVar(&label, "l", "", "Label selector")
	flag.StringVar(&field, "f", "", "Field selector")

	api := clientset.CoreV1()
	// setup list options
	listOptions := metav1.ListOptions{
		LabelSelector: label,
		FieldSelector: field,
	}
	pvcs, err := api.PersistentVolumeClaims(ns).List(listOptions)
	if err != nil {
		log.Fatal(err)
	}
	printPVCs(pvcs)
}

func printPVCs(pvcs *v1.PersistentVolumeClaimList) {
	template := "%-32s%-8s%-8s\n"
	fmt.Printf(template, "NAME", "STATUS", "CAPACITY")
	for _, pvc := range pvcs.Items {
		quant := pvc.Spec.Resources.Requests[v1.ResourceStorage]
		fmt.Printf(
			template,
			pvc.Name,
			string(pvc.Status.Phase),
			quant.String())
	}
}

func GetDeployments(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) ([]v1.Deployment, error) {

	list, err := clientset.AppsV1().Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func encodeDS(clientset *kubernetes.Clientset) {
	ds := &v1.DaemonSet{}
	ds.Name = "example"
	// edit deployment spec

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	enc.Encode(ds)
}

/*
func createDeployment(ctx context.Context, config *rest.Config, ns string) error {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
}
*/
