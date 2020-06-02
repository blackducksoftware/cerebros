NAMESPACE=$1

kubectl create ns $NAMESPACE

SA_NAME=test-sa
kubectl create sa $SA_NAME -n $NAMESPACE
# kubectl create clusterrolebinding test-clusterrole-binding --clusterrole=kube-metrics-clusterrole --user=$SA_NAME
kubectl create clusterrolebinding test-clusterrole-binding --clusterrole=cluster-admin --serviceaccount=$NAMESPACE:$SA_NAME

sed "s/\$NAMESPACE/$NAMESPACE/g" kube-metrics.yaml | \
  sed "s/\$SA_NAME/$SA_NAME/g" | \
  kubectl create -f - -n $NAMESPACE


sleep 10
POD_NAME=$(kubectl get pods -n $NAMESPACE -o json | jq -r '.items[0] | .metadata.name')
kubectl logs -n $NAMESPACE $POD_NAME
