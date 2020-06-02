NAMESPACE=$1

kubectl create ns "$NAMESPACE"

kubectl create -f deploy.yaml -n "$NAMESPACE"
