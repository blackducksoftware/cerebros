FROM ubuntu:16.04

# Git
RUN apt-get update && apt-get install -y git

# Golang
RUN \
    apt-get -y update \
    && apt-get -y install less gdb golang-1.10 \
    && apt-get clean

ENV PATH=/usr/lib/go-1.10/bin:$PATH

# Node
RUN apt-get -y install npm nodejs \
    && npm install -g yarn

# JAVA
RUN apt-get update && \
    apt-get -y install openjdk-8-jdk

# Download and install Gradle
RUN apt-get install -y curl unzip \
    && cd /usr/local \
    && curl -L https://services.gradle.org/distributions/gradle-2.5-bin.zip -o gradle-2.5-bin.zip \
    && unzip gradle-2.5-bin.zip \
    && rm gradle-2.5-bin.zip

# Export some environment variables
ENV GRADLE_HOME=/usr/local/gradle-2.5 PATH=$PATH:$GRADLE_HOME/bin JAVA_HOME=/usr/lib/jvm/java-8-openjdk-amd64

# get maven 3.3.9, verify checksum & install maven
RUN apt-get -y install wget \
    && wget --no-verbose -O /tmp/apache-maven-3.3.9.tar.gz http://archive.apache.org/dist/maven/maven-3/3.3.9/binaries/apache-maven-3.3.9-bin.tar.gz \
    && echo "516923b3955b6035ba6b0a5b031fbd8b /tmp/apache-maven-3.3.9.tar.gz" | md5sum -c \
    && tar xzf /tmp/apache-maven-3.3.9.tar.gz -C /opt/ \
    && ln -s /opt/apache-maven-3.3.9 /opt/maven \
    && ln -s /opt/maven/bin/mvn /usr/local/bin \
    && rm -f /tmp/apache-maven-3.3.9.tar.gz

# Export some environment variables
ENV MAVEN_HOME /opt/maven

# Clang
RUN apt-get install -y clang

# CPAN
RUN apt-get install -qy perl cpanminus \
    && rm -rf "/var/lib/apt/lists/*", "/tmp/*", "/var/tmp/*" \
    && cpanm Proc::ProcessTable Data::Dumper

# remove download archive files
RUN apt-get clean

COPY run-before-scan.sh 'source/run-before-scan.sh'
CMD ["/bin/sh"]
