NAMESPACE=$1
POLARIS_EMAIL=$2
POLARIS_PASSWORD=$3
POLARIS_URL=$4

#kubectl create ns $NAMESPACE

sed "s/\$POLARIS_EMAIL/$POLARIS_EMAIL/g" service-load-gen.yaml | \
  sed "s/\$POLARIS_PASSWORD/$POLARIS_PASSWORD/g" | \
  sed "s/\$POLARIS_URL/$POLARIS_URL/g" | \
  kubectl apply -f - -n $NAMESPACE
