#!/usr/bin/env bash
# Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

###################################
# Prepares guacamole docker image #
###################################

mkdir /var/tmp/guacamole.image
cd /var/tmp/guacamole.image

cat >startup.sh <<-'EOF'
#!/bin/bash

# start up supervisord, all daemons should launched by supervisord.
exec /usr/bin/supervisord -c /opt/safescale/supervisord.conf
EOF

cat >supervisord.conf <<-'EOF'
[supervisord]
nodaemon=true

[unix_http_server]
file=/var/run/supervisor.sock
chmod=0700

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
serverurl=unix:////var/run/supervisor.sock
username=admin
password=admin

[program:guacd]
priority=200
directory=/
command=/usr/sbin/guacd -f
user=root
autostart=true
autorestart=true
stopsignal=QUIT

[program:tomcat]
priority=201
directory=/
command=/usr/local/tomcat/bin/catalina.sh run
user=root
autostart=true
autorestart=true
stopsignal=QUIT
EOF

cat >logback.xml <<-'EOF'
<configuration>
    <!-- Appender for debugging -->
    <appender name="GUAC-DEBUG" class="ch.qos.logback.core.ConsoleAppender">
        <encoder>
            <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>
    <!-- Appender for debugging in a file-->
    <appender name="GUAC-DEBUG_FILE" class="ch.qos.logback.core.FileAppender">
        <file>/var/log/guacd.log</file>
        <encoder>
            <pattern>%d{HH:mm:ss.SSS} [%thread] %-5level %logger{36} - %msg%n</pattern>
        </encoder>
    </appender>
    <!-- Log at DEBUG level -->
    <root level="debug">
        <appender-ref ref="GUAC-DEBUG"/>
        <appender-ref ref="GUAC-DEBUG_FILE"/>
    </root>
</configuration>
EOF

cat >tomcat-users.xml <<-'EOF'
<?xml version='1.0' encoding='utf-8'?>
<tomcat-users>
    <role rolename="admin-gui"/>
    <role rolename="admin-script"/>
    <role rolename="manager-gui"/>
    <role rolename="manager-status"/>
    <role rolename="manager-script"/>
    <role rolename="manager-jmx"/>
    <user name="admin" password="admin" roles="admin-gui,admin-script,manager-gui,manager-status,manager-script,manager-jmx"/>
</tomcat-users>
EOF

cat >Dockerfile <<-'EOF'
FROM debian:sid-slim AS Builder
LABEL maintainer "CS SI"

ARG GUACAMOLE_VERSION=0.9.14
ARG GUACAMOLE_URL=http://apache.org/dyn/closer.cgi?action=download&filename=guacamole/${GUACAMOLE_VERSION}

ENV DEBIAN_FRONTEND noninteractive

# ----------------
# Needed packages
# ----------------
RUN apt-get update -y \
 && apt-get upgrade -y \
 && apt-get install -y \
        build-essential \
        gcc-6 \
        libcairo2-dev \
        libjpeg62-turbo \
        libossp-uuid-dev \
        libpng-dev \
        libvncserver-dev \
        libssh2-1-dev \
        libssl-dev \
        libwebp-dev \
        wget

# Guacamole
WORKDIR /usr/local/src
ADD ${GUACAMOLE_URL}/source/guacamole-server-${GUACAMOLE_VERSION}.tar.gz ./guacamole-server-${GUACAMOLE_VERSION}.tar.gz
RUN tar -zxvf guacamole-server-${GUACAMOLE_VERSION}.tar.gz -C . >/dev/null

RUN cd guacamole-server-${GUACAMOLE_VERSION} \
 && CC=gcc-6 ./configure --prefix=/usr --with-init-dir=/etc/init.d  \
 && make -j3 \
 && make DESTDIR=/usr/local/dist install

#------------------------- DIST phase -------------------------

FROM tomcat:8.5-jre8-slim
LABEL maintainer "CS SI"

ARG GUACAMOLE_VERSION=0.9.14
ARG GUACAMOLE_URL=http://apache.org/dyn/closer.cgi?action=download&filename=guacamole/${GUACAMOLE_VERSION}

ENV DEBIAN_FRONTEND noninteractive

# -----------------
# Needed packages
# -----------------
RUN apt update -y \
 && apt upgrade -y \
 && apt install -y \
        libcairo2 \
        libjpeg62-turbo \
        libossp-uuid16 \
        libpng16-16 \
        libvncclient1 \
        libssh2-1 \
        libssl1.1 \
        libwebp6 \
        procps \
        net-tools \
        supervisor

COPY --from=Builder /usr/local/dist /

# -----------------
# Install Guacamole
# -----------------
# Tomcat
WORKDIR /usr/local/tomcat
RUN rm -rf ./webapps/{examples,doc,ROOT}
ADD ${GUACAMOLE_URL}/binary/guacamole-${GUACAMOLE_VERSION}.war ./webapps/guacamole.war
ADD tomcat-users.xml ./conf/

WORKDIR /root
RUN mkdir .guacamole
ADD logback.xml .guacamole/
ENV GUACAMOLE_HOME /root/.guacamole

RUN mkdir /opt/safescale
WORKDIR /opt/safescale
ADD startup.sh ./
RUN chmod u+x startup.sh
ADD supervisord.conf ./

RUN apt autoremove -y \
 && apt autoclean -y \
 && rm -rf /var/lib/apt/*

# Tomcat Guacamole
EXPOSE 8080

# Tomcat
EXPOSE 8009

ENTRYPOINT ["/opt/safescale/startup.sh"]
EOF

docker build -t guacamole:latest . && \
docker save guacamole:latest | pigz -c9 >/usr/local/dcos/genconf/serve/docker/guacamole.tar.gz && \
cd .. && rm -rf guacamole.image