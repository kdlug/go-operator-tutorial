# GO operator tutorial

Example Memcached operator based on 
Go operator tutorial https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/

## How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Create a dedicated namespace

```shell
kubectl create namespace operator
```

### Switch context

```shell
# using kubens
kubens operator

# using kubectl
kubectl config set-context --current --namespace=operator
```
 
## Running locally

**NOTE:** Make sure that your active is set to `operator`

```shell
# using kubens
$ kubens -c
operator

# using kubectl
$ kubectl config view --minify | grep namespace
    namespace: operator
```
#### Install the CRDs into the cluster:


```sh
make install
```
Verify if CRD is created:

```shell
$ kubectl get crd
NAME                          CREATED AT
memcacheds.cache.github.com   2023-08-18T11:28:05Z
```

#### Run your controller

This will run in the foreground, so switch to a new terminal if you want to leave it running:

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

#### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

#### Create Memcached CR

```shell
kubectl apply -f config/samples/cache_v1alpha1_memcached.yaml
```

Check operator logs

```shell
2023-08-18T14:52:34+02:00	INFO	Creating a new Deployment	{"controller": "memcached", "controllerGroup": "cache.github.com", "controllerKind": "Memcached", "Memcached": {"name":"memcached-sample","namespace":"operator"}, "namespace": "operator", "name": "memcached-sample", "reconcileID": "20046466-8ac8-4488-a0ca-1b4e2f99f7a8", "Deployment.Namespace": "operator", "Deployment.Name": "memcached-sample"}
```

## Cleanup

1. Delete memcached object:

```kubectl 
$ kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml
memcached.cache.github.com "memcached-sample" deleted
```

2. Stop the controller application `ctrl+C`

3. Delete crd

```shell
kubectl delete crd memcacheds.cache.github.com
```


## Running on the cluster
1. Install
l Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/go-operator-tutorial:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/go-operator-tutorial:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```


### Test it out locally

 


## Troubleshooting

### Unable to find MEMCACHED_IMAGE environment variable with the image
```
make run
...
2023-08-18T14:41:50+02:00	ERROR	Reconciler error	{"controller": "memcached", "controllerGroup": "cache.github.com", "controllerKind": "Memcached", "Memcached": {"name":"memcached-sample","namespace":"operator"}, "namespace": "operator", "name": "memcached-sample", "reconcileID": "c1fa0f69-4b37-42ae-a887-6e17e045c873", "error": "Unable to find MEMCACHED_IMAGE environment variable with the image"}
```

You have to provide image for deployment using MEMCACHED_IMAGE variable.

F.ex.

```shell
make MEMCACHED_IMAGE="memcached:1.4.36-alpine"  run
```

### Unable to delete CR

When the following command got stuck:
```shell
$ kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml
```

Edit CR:

```shell
$ kubectl edit memcached memcached-sample
```

Remove sections:
- deletionTimestamp:
- finalizers:

Run delete command again:

```shell
$ kubectl delete -f config/samples/cache_v1alpha1_memcached.yaml
```
