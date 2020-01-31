FROM scratch

COPY ./runPolarisJob ./runPolarisJob

COPY /Users/hammer/go/src/github.com/blackducksoftware/JavaVulnerableLab ./data/JavaVulnerableLab
COPY /Users/hammer/go/src/github.com/blackducksoftware/hub-fortify-parser ./data/hub-fortify-parser
COPY /Users/hammer/go/src/github.com/blackducksoftware/rabbitmq ./data/rabbitmq
COPY /Users/hammer/go/src/github.com/blackducksoftware/polaris-deploy-sanity ./data/polaris-deploy-sanity
COPY /Users/hammer/go/src/github.com/blackducksoftware/synopsys-operator ./data/synopsys-operator

CMD ["./runPolarisJob"]
