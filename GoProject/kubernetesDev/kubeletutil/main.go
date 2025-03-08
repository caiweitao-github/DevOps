package kubeletutil

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	"util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var log = util.NewInitLog("/data/kdl/log/kubeServer/kubeletutil.log")

var url = ""

var title = "[k8s镜像自动部署通知]"

var imageBase = ""

var clientset *kubernetes.Clientset

type KdlConfig struct {
	Ns         string `json:"ns"`
	DeployName string `json:"deploy"`
	Image      string `json:"image"`
}

func init() {
	var err error
	home := homedir.HomeDir()
	if home == "" {
		panic("get home dir fail.")
	}
	kubeconfig := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic("create kubernets client err: " + err.Error())
	}
}

func (k *KdlConfig) UpdateDeploy() error {
	if k.DeployName == "resdock-cab" {
		k.UpdateDeployLabel()
		return nil
	}
	deploymentsClient := clientset.AppsV1().Deployments(k.Ns)
	deployment, err := deploymentsClient.Get(context.TODO(), k.DeployName, metav1.GetOptions{})
	if err != nil {
		util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: %s", k.DeployName, err.Error())})
		return err
	}
	oldImageVersion := deployment.Spec.Template.Spec.Containers[0].Image
	if oldImageVersion == imageBase+k.Image {
		k.RestartDeployment()
		return nil
	}
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Spec.Template.Spec.Containers[0].Image = imageBase + k.Image
	} else {
		util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: 容器数量小于0.", k.DeployName)})
	}

	_, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: %s", k.DeployName, err.Error())})
		return err
	}
	util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s 镜像版本: %s -> %s\n", k.DeployName, oldImageVersion, imageBase+k.Image)})
	return nil
}

func (k *KdlConfig) UpdateDeployLabel() error {
	deploymentsClient := clientset.AppsV1().Deployments(k.Ns)
	labelSelector := "k8s-app=resdock-cab"
	deployments, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: %s", k.DeployName, err.Error())})
		return err
	}

	if len(deployments.Items) == 0 {
		util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: 为找到符合条件的Deployment", k.DeployName)})
		return err
	}

	for _, deployment := range deployments.Items {
		if len(deployment.Spec.Template.Spec.Containers) == 0 {
			continue
		}
		for i := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Image = imageBase + k.Image
		}
		_, err = deploymentsClient.Update(context.TODO(), &deployment, metav1.UpdateOptions{})
		if err != nil {
			util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: %s", deployment.Name, err.Error())})
			continue
		}
	}
	util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s 镜像版本: %s\n", k.DeployName, imageBase+k.Image)})
	return nil
}

func (k *KdlConfig) RestartDeployment() error {
	deploymentsClient := clientset.AppsV1().Deployments(k.Ns)
	deployment, err := deploymentsClient.Get(context.TODO(), k.DeployName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	_, err = deploymentsClient.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		util.FeiShuNotify(url, title, []string{fmt.Sprintf("%s部署失败，原因: %s", k.DeployName, err.Error())})
		return err
	}
	util.FeiShuNotify(url, title, []string{fmt.Sprintf("相同的镜像版本，不更新，自动重启，%s 镜像版本: %s\n", k.DeployName, deployment.Spec.Template.Spec.Containers[0].Image)})
	return nil
}
