NAMESPACE=$1
POLARIS_EMAIL=$2
POLARIS_PASSWORD=$2
POLARIS_IP=$4

kubectl create ns $NAMESPACE

sed "s/\$POLARIS_EMAIL/$POLARIS_EMAIL/g" polaris-cli.yaml | \
  sed "s/\$POLARIS_PASSWORD/$POLARIS_PASSWORD/g" | \
  sed "s/\$POLARIS_IP/$POLARIS_IP/g" | \
  kubectl create -f - -n $NAMESPACE

kubectl create -f scan-queue.yaml -n $NAMESPACE
