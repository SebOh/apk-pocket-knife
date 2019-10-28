#####
## BUILD THE CFR TOOL
#####

FROM maven:3.6.1-jdk-8-slim AS CFR_BUILDER

RUN curl -L https://github.com/leibnitz27/cfr/archive/0.146.tar.gz --output 0.146.tar.gz && \
    tar xzf 0.146.tar.gz && \
    rm 0.146.tar.gz && \
    cd cfr-0.146 && \
    sed -i 's/<additionalOptions>-html5 -quiet<\/additionalOptions>//g' pom.xml &&\
    mvn package


#####
## BUILD THE VDEX EXTRACTOR TOOL
#####

FROM alpine:3.10.2 AS VDEX_EXTRACTOR

RUN wget --quiet https://github.com/anestisb/vdexExtractor/archive/0.5.2.zip && \
    unzip 0.5.2.zip && \
    mv vdexExtractor-0.5.2/ vdex && \
    rm 0.5.2.zip && \
    cd vdex && \
    apk -U add make gcc zlib-dev musl-dev bash && \
    bash && \
    ./make.sh

#####
## BUILD THE ENTRY POINT FOR GO
#####

FROM golang:1.13.2-alpine3.10 AS GO_BUILDER

COPY entrypoint.go /go/src/app/entrypoint.go

RUN cd /go/src/app/ && \
    go build entrypoint.go

#####
## MAIN BUILD
#####

FROM azul/zulu-openjdk-alpine:11.0.4-jre

LABEL maintainer="sebastian_ohm@gmx.net" \
      version="2.0.0"

ENV PATH="$PATH:$JAVA_HOME/bin"

ENV TOOLS /tools
ENV APK_DIR /apk
ENV SCRIPTS /script

# Install Packages

RUN apk update && \
    apk add jq bash &&\
    mkdir ${TOOLS}

# Install CFR

COPY --from=CFR_BUILDER /cfr-0.146/target/cfr-0.146.jar ${TOOLS}/cfr/cfr.jar

ENV CFR_HOME=${TOOLS}/cfr
ENV PATH="$PATH:$CFR_HOME"

RUN chmod +x $CFR_HOME/cfr.jar

# Install VDEX Extractor

COPY --from=VDEX_EXTRACTOR /vdex/bin/vdexExtractor ${TOOLS}/vdex/vdexExtractor

ENV VDEX_HOME=${TOOLS}/vdex
ENV PATH="$PATH:$VDEX_HOME"

RUN chmod +x $VDEX_HOME/vdexExtractor

# Install JADX

WORKDIR ${TOOLS}

RUN wget --quiet https://github.com/skylot/jadx/releases/download/v1.0.0/jadx-1.0.0.zip && \
    mkdir jadx && \
    unzip -qq jadx-1.0.0.zip -d jadx && \
    rm jadx-1.0.0.zip

ENV JADX_HOME=${TOOLS}/jadx
ENV PATH="$PATH:$JADX_HOME/bin"

# Install APKTOOL

ENV APK_TOOL_HOME=${TOOLS}/apktool
ENV APK_TOOL_VERSION=2.4.0
ENV PATH="$PATH:$APK_TOOL_HOME"

RUN mkdir ${APK_TOOL_HOME} && \
    cd ${APK_TOOL_HOME} && \
    JAR_APK_TOOL=$(wget -qO - https://api.bitbucket.org/2.0/repositories/iBotPeaches/apktool/downloads | jq -c '.values[] | {name: .name, link: .links.self.href}' | jq --arg version "$APK_TOOL_VERSION" -r '. | select(.name | contains($version)).link') && \
    echo ${JAR_APK_TOOL} && \
    wget --quiet ${JAR_APK_TOOL} && \
    mv "apktool_${APK_TOOL_VERSION}.jar" apktool.jar && \
    wget --quiet https://raw.githubusercontent.com/iBotPeaches/Apktool/master/scripts/linux/apktool && \
    chmod +x apktool.jar && \
    chmod +x apktool

# Install dex2Jar

RUN wget --quiet https://bitbucket.org/pxb1988/dex2jar/downloads/dex2jar-2.0.zip && \
    mkdir dex2Jar && \
    unzip -qq dex2jar-2.0.zip -d dex2Jar && \
    rm dex2jar-2.0.zip && \
    cd dex2Jar/dex2jar-2.0 && \
    find *.sh -type f -exec chmod +x {} \;

ENV DEX2JAR_HOME=${TOOLS}/dex2Jar
ENV PATH="$PATH:$DEX2JAR_HOME/dex2jar-2.0"

# Install Procyon

RUN wget --quiet https://bitbucket.org/mstrobel/procyon/downloads/procyon-decompiler-0.5.36.jar && \
    mkdir procyon && \
    mv procyon-decompiler-0.5.36.jar procyon/ &&\
    chmod +x procyon/procyon-decompiler-0.5.36.jar

ENV PROCYON_HOME=${TOOLS}/procyon
ENV PATH="$PATH:$PROCYON_HOME"

# Prepare Auto Decompile

WORKDIR ${SCRIPTS}

COPY --from=GO_BUILDER /go/src/app/entrypoint entrypoint
COPY decompile.sh decompile
RUN chmod +x decompile

WORKDIR ${APK_DIR}
ENTRYPOINT [".././script/decompile"]