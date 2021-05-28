package main

import (
	"encoding/json"
	"fmt"
//	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"os"
	"time"
//        "path"
	"k8s.io/client-go/kubernetes"
	//        "context"
	"serviceDeploy/k8s"
	//"k8s.io/client-go/util/retry"
	"net/http"
)

type basicAuthRoundTripper struct {
	headers map[string][]string
	rt http.RoundTripper
}

func (basicAuth basicAuthRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	for key, values := range basicAuth.headers {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	return basicAuth.rt.RoundTrip(request)
}

//监听Deployment变化
func startWatchDeployment(deploymentsClient v1.DeploymentInterface) {
	w, _ := deploymentsClient.Watch(metav1.ListOptions{})
	for {
		select {
		case e, _ := <-w.ResultChan():
			fmt.Println(e.Type, e.Object)
		}
	}
}

func verify(deployment_client v1.DeploymentInterface, deployment_name string,imageName string)bool{

	var realContainerImage string
	var returnValue bool = false
	d,err := deployment_client.Get(deployment_name,metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	containCounts := len(d.Spec.Template.Spec.Containers)

	for i := 0; i < containCounts; i++ {
		if d.Spec.Template.Spec.Containers[i].Name == deployment_name{
			realContainerImage = d.Spec.Template.Spec.Containers[i].Image
			var statusAvailableReplicas = d.Status.AvailableReplicas
            var specAvailableReplicas = d.Spec.Replicas
            var statusReplicas = d.Status.Replicas
            var specReplicas = d.Spec.Replicas
            var statusUpdatedReplicas = d.Status.UpdatedReplicas
            var specUpdatedReplicas = d.Spec.Replicas
            var statusObservedGeneration = d.Status.ObservedGeneration
            var specGeneration = d.Generation
			if realContainerImage == imageName && (statusAvailableReplicas == *specAvailableReplicas) &&
            	(statusReplicas == *specReplicas) && (statusUpdatedReplicas == * specUpdatedReplicas) &&
				(statusObservedGeneration >= specGeneration){
				fmt.Println("pass")
				returnValue = true
			}
			break
		}
	}
	//fmt.Println("****************************")
	//fmt.Println(deployment_name)
	//fmt.Println(realContainerImage)
	//fmt.Println("****************************")
	return  returnValue
}

func getDeploymentResult(deployment_client v1.DeploymentInterface, deployment_name string,imageName string)bool{
	var returnValue = false
	var retryTotalCounts = 6
	for i := 0; i < retryTotalCounts ;i++{
		if ! verify(deployment_client, deployment_name,imageName) {
			time.Sleep(time.Duration(10)*time.Second)
		}else{
			returnValue = true
			break
		}
	}
return returnValue
}

func main() {

	//传参
	if len(os.Args[1:]) != 4{
		fmt.Printf("Arguments counts must be 4!\nUsage:./serviceDeploy <namepsace> <deploymentName> " +
			"<serviceName> <imageName> \n")
		os.Exit(1)
	}
	nameSpace := os.Args[1]
	deploymentName := os.Args[2]
	containerName := os.Args[3]
	imageName := os.Args[4]
	//retry.RetryOnConflict()
	//1.读取配置信息
        //dir,_:= os.Getwd()
	//fmt.Println(dir)
        //configFile := path.Join(dir,"config")
/*
	err := k8s.InitConfig("config")
	if err != nil {
		fmt.Printf("Error reading configuration: %s\n", err.Error())
		os.Exit(1)
	}
*/
/*
	iamUrl := viper.GetString("iam_config.url")
	acount := viper.GetString("iam_config.account")
	secret := viper.GetString("iam_config.secret")
	appId := viper.GetString("Native_API.appId")
	env := viper.GetString("Native_API.envCode")
	userId := viper.GetString("Native_API.userId")
	apiUrl := fmt.Sprintf("%s/%s", viper.GetString("Native_API.url"), viper.GetString("Native_API.clusterId"))
*/
        iamUrl := "http://iam.hics.huawei.com/internal/so/iam/tokens"
        acount := "com.huawei.group1569726916857@com.huawei.group1569726916857" 
        secret := "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        appId := "com.huawei.group1569726916857"
        env := "pro_biz_rnd_01"
        userId := "cwx1024877"
        apiUrl :=  "http://hics-dgg-asc001.huawei.com/internal/iaas/compute/ecs/ecstnv2/clusters/665a9f83-54f2-49ef-b933-64acb2f5e924"
	//2.获取鉴权token
	token, err := k8s.GetAuthorization(iamUrl,acount,secret,appId,env,userId)
	//fmt.Println("token: ", token)

	if err != nil {
		panic(err.Error())
	}

	//3.原生网关鉴权需要自己定义一个config，用于生成Clientset，Clientset是操作所有资源的句柄
	config := rest.Config{
		Host: apiUrl,
		TLSClientConfig:rest.TLSClientConfig{Insecure: true},
		WrapTransport: func(rt http.RoundTripper) (http.RoundTripper){
			return &basicAuthRoundTripper{
				headers: map[string][]string{
					"Authorization":  []string{fmt.Sprintf("Bearer %s", token)},
					//"ContentType":  []string{"application/json"},
				},
				rt: rt,
			}
		},
	}

	//4.根据定义的rest.Config创建clientset
	clientset, err := kubernetes.NewForConfig(&config)
	if err != nil {
		panic(err.Error())
	}
	//clientset.ExtensionsV1beta1().
	patchData:=map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]string{{
						"image": imageName,
						"name": containerName,
					},
					},
				},
			},
		},
	}
	patchByte,_:=json.Marshal(patchData)
	println(string(patchByte))
	deploymentsClient := clientset.AppsV1().Deployments(nameSpace)
	_,err = deploymentsClient.Patch(deploymentName,types.StrategicMergePatchType,patchByte)
	//fmt.Println(nodes.Items)
	if err != nil {
		panic(err.Error())
	}
	//      deploymentResponse, err1 := deploymentsClient.Get(deploymentName,metav1.GetOptions{})
	//startWatchDeployment(deploymentsClient)
	//      fmt.Println("deployment query result: ", *deploymentResponse)
	/*
	   if err1 != nil {
	           panic(err.Error())
	   }
	*/
	if ! getDeploymentResult(deploymentsClient,deploymentName,imageName){
fmt.Println("fail")       
os.Exit(1)
	}
}
