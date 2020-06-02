import subprocess
import sys


deployments = {
	"auth-server": 1,
	"configs-service": 1,
	"cos-server": 1,
	"desktop-metrics-service": 1,
	"download-minio": 1,
	"issue-server": 1,
	"jobfarmautoscaler": 1,
	"jobs-controller-service": 1,
	"jobs-service": 1,
	"logs-service": 1,
	
	# Don't scale down mongodb!  It won't be able to come back up!
	# "mongodb": 1,
	
	"notifications-service": 1,
	"pericles-swagger-ui": 1,
	"polaris-db-vault": 1,
	"postgresql": 1,
	"taxonomy-server": 1,
	"tds-code-analysis": 1,
	"tools-service": 1,
	"triage-command-handler": 1,
	"triage-query": 1,
	"upload-minio": 1,
	"vault-init": 1,
	"vinyl-server": 1,
	"web-core": 1,
	"web-help": 1
}

statefulsets = {
	"eventstore": 3,
	"polaris-db-consul": 1
}


def scale(action, ns):
	for (d, replicas) in deployments.items():
		count = 0 if action == "down" else replicas
		command = ["kubectl", "scale", "deployment", d, "--replicas={}".format(count), "-n", ns]
		print("about to run: <{}>".format(" ".join(command)))
		proc = subprocess.Popen(command,
			stdout=subprocess.PIPE, 
			stderr=subprocess.PIPE)
		stdout, stderr = proc.communicate()
		print("stdout and stderr: \n{}\n{}\n".format(stdout, stderr))
	for (s, replicas) in statefulsets.items():
		count = 0 if action == "down" else replicas
		command = ["kubectl", "scale", "statefulset", s, "--replicas={}".format(count), "-n", ns]
		print("about to run: <{}>".format(" ".join(command)))
		proc = subprocess.Popen(command,
			stdout=subprocess.PIPE, 
			stderr=subprocess.PIPE)
		stdout, stderr = proc.communicate()
		print("stdout and stderr: \n{}\n{}\n".format(stdout, stderr))


action, ns = sys.argv[1:3]
scale(action, ns)
