NAMESPACE=$1
BD_USER=$2
BD_PASSWORD=$3
BD_HOST=$4

kubectl create ns "$NAMESPACE"

helm install my-fave-prom stable/prometheus --namespace $NAMESPACE

SA_NAME=bd-kube-metrics-sa
kubectl create sa $SA_NAME -n $NAMESPACE
kubectl create clusterrolebinding bd-kube-metrics-clusterrole-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=$NAMESPACE:$SA_NAME

COLLECTION_NAMESPACE="$NAMESPACE"
sed "s/\$COLLECTION_NAMESPACE/$COLLECTION_NAMESPACE/g" kube-metrics.yaml | \
  sed "s/\$SA_NAME/$SA_NAME/g" | \
  kubectl create -f - -n $NAMESPACE

sed "s/\$BD_USER/$BD_USER/g" blackduck-cli.yaml | \
  sed "s/\$BD_PASSWORD/$BD_PASSWORD/g" | \
  sed "s/\$BD_PASSWORD/$BD_PASSWORD/g" | \
  sed "s/\$BD_HOST/$BD_HOST/g" | \
  kubectl create -n "$NAMESPACE" -f -

kubectl create -f scan-queue.yaml -n "$NAMESPACE"
