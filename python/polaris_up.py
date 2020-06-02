import re
import json
import subprocess


repos_text = [
  "https://sig-gitlab.internal.synopsys.com/blackduck/hub",
  "https://sig-gitlab.internal.synopsys.com/blackduck/hub",
  "https://sig-gitlab.internal.synopsys.com/swip/bootstrap",
  "https://sig-gitlab.internal.synopsys.com/swip/ngp",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/swip/ngp",
  "https://sig-gitlab.internal.synopsys.com/swip/jobs",
  "https://sig-gitlab.internal.synopsys.com/swip/swip",
  "https://sig-gitlab.internal.synopsys.com/swip/env",
  "https://sig-gitlab.internal.synopsys.com/swip/swagger",
  "https://sig-gitlab.internal.synopsys.com/swip/ngp",
  "https://sig-gitlab.internal.synopsys.com/swip/logs",
  "https://sig-gitlab.internal.synopsys.com/swip/tools",
  "https://sig-gitlab.internal.synopsys.com/swip/notifications",
  "https://sig-gitlab.internal.synopsys.com/swip/jira",
  "https://sig-gitlab.internal.synopsys.com/swip/configs",
  "https://sig-gitlab.internal.synopsys.com/swip/polaris",
  "https://sig-gitlab.internal.synopsys.com/swip/polaris",
  "https://sig-gitlab.internal.synopsys.com/swip/cli",
  "https://sig-gitlab.internal.synopsys.com/swip/web",
  "https://sig-gitlab.internal.synopsys.com/swip/swip",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/swip/pericles",
  "https://sig-gitlab.internal.synopsys.com/clops/polaris",
  "https://sig-gitlab.internal.synopsys.com/clops/swip",
  "https://sig-gitlab.internal.synopsys.com/swip/polaris",
  "https://sig-gitlab.internal.synopsys.com/swip/jsonapi",
  "https://sig-gitlab.internal.synopsys.com/swip/event",
  "https://sig-gitlab.internal.synopsys.com/swip/prevent",
  "https://sig-gitlab.internal.synopsys.com/cim/issue"
]


repos = [
  [
    "https://sig-gitlab.internal.synopsys.com/protecode-sc/helm-chart",
    "protecode-sc/helm-chart"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/bootstrap",
    "swip/bootstrap"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/ngp-maven-parent",
    "swip/ngp-maven-parent"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-devenv",
    "swip/pericles-devenv"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/ngp-pipeline",
    "swip/ngp-pipeline"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/jobs-service",
    "swip/jobs-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/swip-backend",
    "swip/swip-backend"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/env-scripting",
    "swip/env-scripting"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/swagger-maven-plugin-ext",
    "swip/swagger-maven-plugin-ext"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/ngp-bootstrap",
    "swip/ngp-bootstrap"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/logs-service",
    "swip/logs-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/tools-service",
    "swip/tools-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/notifications-service",
    "swip/notifications-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/jira-service",
    "swip/jira-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/configs-service",
    "swip/configs-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/polaris-vault-lib",
    "swip/polaris-vault-lib"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/polaris-crypto",
    "swip/polaris-crypto"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/cli",
    "swip/cli"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/web-help",
    "swip/web-help"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/swip-frontend",
    "swip/swip-frontend"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-tds-sca",
    "swip/pericles-tds-sca"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-tds-seeker",
    "swip/pericles-tds-seeker"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-tds-code-analysis",
    "swip/pericles-tds-code-analysis"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-tds-fuzz",
    "swip/pericles-tds-fuzz"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-tds-csv",
    "swip/pericles-tds-csv"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/pericles-swagger-ui",
    "swip/pericles-swagger-ui"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-common-lib",
    "reporting-platform/rp-common-lib"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/tool-common-lib",
    "reporting-platform/tool-common-lib"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-seeker-agent",
    "reporting-platform/rp-seeker-agent"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/ms-portal-agent",
    "reporting-platform/ms-portal-agent"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/black-duck-agent",
    "reporting-platform/black-duck-agent"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/cim-agent",
    "reporting-platform/cim-agent"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-frontend",
    "reporting-platform/rp-frontend"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-swagger-doc",
    "reporting-platform/rp-swagger-doc"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-polaris-agent-service",
    "reporting-platform/rp-polaris-agent-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-issue-manager",
    "reporting-platform/rp-issue-manager"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-report-service",
    "reporting-platform/rp-report-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-portfolio-service",
    "reporting-platform/rp-portfolio-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-tools-portfolio-service",
    "reporting-platform/rp-tools-portfolio-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/sig-desktop/desktop-metrics-service",
    "sig-desktop/desktop-metrics-service"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/clops/polaris-build-common",
    "clops/polaris-build-common"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/clops/swip-helmcharts-app",
    "clops/swip-helmcharts-app"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/polaris-db-test-lib",
    "swip/polaris-db-test-lib"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/jsonapi-client",
    "swip/jsonapi-client"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/event-stream-migration",
    "swip/event-stream-migration"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/swip/prevent-error-java-lib",
    "swip/prevent-error-java-lib"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/cim/issue-type-taxonomies",
    "cim/issue-type-taxonomies"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-i18n-utils",
    "reporting-platform/rp-i18n-utils"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-owasp-utils",
    "reporting-platform/rp-owasp-utils"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-rest-template",
    "reporting-platform/rp-rest-template"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-agent-rest-client",
    "reporting-platform/rp-agent-rest-client"
  ],
  [
    "https://sig-gitlab.internal.synopsys.com/reporting-platform/rp-auth-filter-lib",
    "reporting-platform/rp-auth-filter-lib"
  ]
]

# git@sig-gitlab.internal.synopsys.com:blackduck/hub/hub-backend.git

def parse_repos_from_chrome_bookmarks(bookmarks_path):
	pattern = re.compile('(https://sig-gitlab.internal.synopsys.com/([^"]+))')
	with open(bookmarks_path) as links:
		out = pattern.findall(links.read())
		# print("\n".join("\t".join([g[0], g[1], g[2]]) for g in out))
		# print(type(out))
		print(json.dumps(out, indent=2))


def clone_repos():
	for (repo, name) in repos:
		fixed_repo = "git@sig-gitlab.internal.synopsys.com:{}.git".format(name)
		command = ['git', 'clone', fixed_repo, 'polaris-{}'.format(name.replace("/", "-"))]
		print("about to run: {}".format(" ".join(command)))
		proc = subprocess.Popen(command,
			stdout=subprocess.PIPE, 
			stderr=subprocess.PIPE)
		stdout, stderr = proc.communicate()
		print(stderr)
		print(stdout)
		print("\n")


clone_repos()
