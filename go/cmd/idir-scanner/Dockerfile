FROM ubuntu

RUN apt-get update && apt-get install -y ca-certificates

COPY build/idir_scanner_linux ./runPolarisJob

CMD ["./runPolarisJob"]
