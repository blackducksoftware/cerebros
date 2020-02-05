FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates

COPY build/polaris /usr/local/bin/polaris
COPY build/polaris-submit /usr/local/bin/polaris-submit
COPY build/polaris-wait /usr/local/bin/polaris-wait
COPY build/idir_scanner_linux ./runPolarisJob

CMD ["./runPolarisJob"]
