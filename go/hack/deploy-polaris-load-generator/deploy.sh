NAMESPACE=$1
POLARIS_IP=$2
POLARIS_EMAIL=$3
POLARIS_PASSWORD=$4
POLARIS_HOST=$5 # onprem-perf.dev.polaris.synopsys.com

kubectl create ns $NAMESPACE

helm install my-fave-prom stable/prometheus --namespace $NAMESPACE

SA_NAME=polaris-load-gen-kube-metrics-sa
kubectl create sa $SA_NAME -n $NAMESPACE
kubectl create clusterrolebinding kube-metrics-clusterrole-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=$NAMESPACE:$SA_NAME

sed "s/\$POLARIS_PASSWORD/$POLARIS_PASSWORD/g" polaris-cli.yaml | \
  sed "s/\$POLARIS_IP/$POLARIS_IP/g" | \
  sed "s/\$POLARIS_EMAIL/$POLARIS_EMAIL/g" | \
  sed "s/\$POLARIS_HOST/$POLARIS_HOST/g" | \
  kubectl create -f - -n $NAMESPACE

kubectl create -f scan-queue.yaml -n $NAMESPACE

sed "s/\$POLARIS_PASSWORD/$POLARIS_PASSWORD/g" polaris-api-load-gen.yaml | \
  sed "s/\$POLARIS_IP/$POLARIS_IP/g" | \
  sed "s/\$POLARIS_EMAIL/$POLARIS_EMAIL/g" | \
  sed "s/\$POLARIS_HOST/$POLARIS_HOST/g" | \
  kubectl create -f - -n $NAMESPACE

COLLECTION_NAMESPACE="$NAMESPACE"
sed "s/\$COLLECTION_NAMESPACE/$COLLECTION_NAMESPACE/g" kube-metrics.yaml | \
  sed "s/\$SA_NAME/$SA_NAME/g" | \
  kubectl create -f - -n $NAMESPACE


export POD_NAME=$(kubectl get pods --namespace $NAMESPACE -l "app=prometheus,component=server" -o jsonpath="{.items[0].metadata.name}")
kubectl --namespace $NAMESPACE port-forward $POD_NAME 9090

# add these to pods' metadata in order to get prometheus to scrape them:
#metadata:
#  annotations:
#    prometheus.io/scrape: "true"
#    prometheus.io/path: /metrics
#    prometheus.io/port: "8080"

