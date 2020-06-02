NS=$1

helm delete my-fave-prom

kubectl delete clusterrolebinding kube-metrics-clusterrole-binding

kubectl delete ns "$NS"
