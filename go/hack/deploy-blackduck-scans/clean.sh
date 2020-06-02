NS=$1

helm uninstall my-fave-prom --namespace $NS

kubectl delete clusterrolebinding bd-kube-metrics-clusterrole-binding

kubectl delete ns "$NS"
