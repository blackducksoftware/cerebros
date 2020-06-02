NS=$1

kubectl delete -f service-load-gen.yaml -n $NS
