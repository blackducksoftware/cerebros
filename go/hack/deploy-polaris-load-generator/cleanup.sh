NS=$1

#helm delete my-fave-prom --namespace $NS
helm uninstall my-fave-prom --namespace $NS

kubectl delete clusterrolebinding kube-metrics-clusterrole-binding

kubectl delete clusterrole my-fave-prom-kube-state-metrics
kubectl delete clusterrole my-fave-prom-prometheus-alertmanager
kubectl delete clusterrole my-fave-prom-prometheus-pushgateway
kubectl delete clusterrole my-fave-prom-prometheus-server

kubectl delete clusterrolebindings.rbac.authorization.k8s.io my-fave-prom-kube-state-metrics
kubectl delete clusterrolebindings.rbac.authorization.k8s.io my-fave-prom-prometheus-alertmanager
kubectl delete clusterrolebindings.rbac.authorization.k8s.io my-fave-prom-prometheus-pushgateway
kubectl delete clusterrolebindings.rbac.authorization.k8s.io my-fave-prom-prometheus-server

kubectl delete ns "$NS"
