package main

import (
	"flag"
	"log"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/cache"


	"github.com/audunstrand/deployer/tpr"

	// Only required to authenticate against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"time"

)

var (
	config *rest.Config
)

func main() {
	log.Println("starting to deploy")
	defer log.Println("done deploying")

	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()


	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	// initialize third party resource if it does not exist
	err = initTpr(err, clientset)

	// make a new config for our extension's API group, using the first config as a baseline
	var tprconfig *rest.Config
	tprconfig = config
	configureClient(tprconfig)

	/*tprclient, err := rest.RESTClientFor(tprconfig)
	if err != nil {
		panic(err)
	}

	var app tpr.App
	*/
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "app", "default", nil)

	for {

		_, controller := cache.NewInformer(watchlist, &tpr.AppList{}, time.Second*2, cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Printf("add: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				log.Printf("delete: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Printf("old: %s, new: %s \n", oldObj, newObj)
			},
		})

		stop := make(chan struct{})
		go controller.Run(stop)
		time.Sleep(time.Second * 5)
	}
	/*err = tprclient.Get().
		Resource("apps").
		Namespace(api.NamespaceDefault).
		Name("app2").
		Do().Into(&app)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create an instance of our TPR
			example := &tpr.App{
				Metadata: api.ObjectMeta{
					Name: "app2",
				},
				Spec: tpr.AppSpec{
					Name:   "app2",
					Active: true,
				},
			}

			var result tpr.App
			err = tprclient.Post().
				Resource("apps").
				Namespace(api.NamespaceDefault).
				Body(example).
				Do().Into(&result)

			if err != nil {
				panic(err)
			}
			log.Printf("CREATED: %#v\n", result)
		} else {
			panic(err)
		}
	} else {
		log.Printf("GET: %#v\n", app)
	}

	// Fetch a list of our TPRs

	appList := tpr.AppList{}
	err = tprclient.Get().Resource("apps").Do().Into(&appList)
	if err != nil {
		panic(err)
	}
	for _, app := range appList.Items {
		log.Println(app.Spec.Name)
	}*/

}
func initTpr(err error, clientset *kubernetes.Clientset) error {
	thpr, err := clientset.Extensions().ThirdPartyResources().Get("app.k8s.io")
	if err != nil {
		if errors.IsNotFound(err) {
			thpr := &v1beta1.ThirdPartyResource{
				ObjectMeta: v1.ObjectMeta{
					Name: "app.k8s.io",
				},
				Versions: []v1beta1.APIVersion{
					{Name: "v1"},
				},
				Description: "An App ThirdPartyResource",
			}

			result, err := clientset.Extensions().ThirdPartyResources().Create(thpr)
			if err != nil {
				panic(err)
			}
			log.Printf("CREATED: %#v\nFROM: %#v\n", result, thpr)
		} else {
			panic(err)
		}
	} else {
		log.Printf("SKIPPING: already exists %#v\n", thpr.Name)
	}
	return err
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func configureClient(config *rest.Config) {
	groupversion := unversioned.GroupVersion{
		Group:   "k8s.io",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}

	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				&tpr.App{},
				&tpr.AppList{},
				&api.ListOptions{},
				&api.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(api.Scheme)
}

func Map(vs []tpr.App, f func(app tpr.App) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
