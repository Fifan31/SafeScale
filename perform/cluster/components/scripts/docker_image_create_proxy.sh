#!/usr/bin/env bash
#
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

########################################
# Prepares reverse proxy for guacamole #
########################################

mkdir /tmp/proxy.image

cat >/tmp/proxy.image/startup.sh <<-EOF
#!/bin/bash

function update_file {
    # Take two files in arguments : first must be docker default conf file, second must be the file used
    # If file $1 is newer than file $2
    if [ -f $2 ] && [ $1 -nt $2 ]
    then
        CHECKSUM_DOCKER_FILE=`md5sum $1 | tr -s ' ' | cut -d ' ' -f1`
        CHECKSUM_CONF_FILE=`md5sum $2 | tr -s ' ' | cut -d ' ' -f1`
        if [ "${CHECKSUM_DOCKER_FILE}" != "${CHECKSUM_CONF_FILE}" ]
        then
            # File has been updated => we save the old conf file and replace it by docker
            DATE=`date +%Y-%m-%d-%H-%M-%S`
            mv $2 $2-${DATE}.confsave
            cp $1 $2
        fi
    else
        if [ ! -f $2 ]
        then
            # File doesn't exist => we create it from default conf file
            cp $1 $2
        fi
    fi
}

function update_conf {
    # Take two folders in arguments : first must be docker default conf folder, second must be used folder containing same conf files
    for file in `ls $1`
    do
        if [ -f $1/${file} ]
        then
            update_file $1/${file} $2/${file}
        fi
    done
}

# Path to default conf stored inside docker during build
DATA_DOCKER_CONF=/data/docker-conf
# Update conf file (only if conf file stored during build is more recent than current used file)
update_conf ${DATA_DOCKER_CONF}/apache2-conf/ /apache2-conf/
update_conf ${DATA_DOCKER_CONF}/Key/ /certificate/
update_conf ${DATA_DOCKER_CONF}/logrotate.d/ /etc/logrotate.d/
update_conf ${DATA_DOCKER_CONF}/sites-available/ /etc/apache2/sites-available/

# If needed we change conf using requested domain name
if [ ! -z ${DOMAIN_NAME+x} ] && [ "${DOMAIN_NAME}" != "" ]
then
    echo "Starting proxy on domain : ${DOMAIN_NAME}"
    # Create all needed files
    if [ "${DOMAIN_NAME}" != "${DEFAULT_DOMAIN_NAME}" ]
    then
        #Rename apache conf files
        update_file ${DATA_DOCKER_CONF}/sites-available/${DEFAULT_DOMAIN_NAME}.conf /etc/apache2/sites-available/${DOMAIN_NAME}.conf

    fi
else
    echo "Starting proxy on default domain : ${DEFAULT_DOMAIN_NAME}"
    DOMAIN_NAME=${DEFAULT_DOMAIN_NAME}
fi

# Replace template tags by domain name
sed -i -e "s#%%DOMAIN_NAME%%#${DOMAIN_NAME}#g" /etc/apache2/sites-available/000-default.conf
sed -i -e "s#%%DOMAIN_NAME%%#${DOMAIN_NAME}#g" /etc/apache2/sites-available/${DOMAIN_NAME}.conf

a2dissite ${DEFAULT_DOMAIN_NAME}.conf
a2ensite ${DOMAIN_NAME}.conf

# Make sure Apache will start no matter what.
rm -f /var/run/apache2/apache2.pid &>/dev/null

# start up supervisord, all daemons should launched by supervisord.
exec /usr/bin/supervisord -c /opt/supervisord.conf
EOF

cat >/tmp/proxy.image/supervisord.conf <<-EOF
[supervisord]
nodaemon=true
logfile=/var/log/supervisord.log
# With log level debug, the supervisord log file will record the stderr/stdout
# output of its child processes and extended info info about process state
# changes
loglevel=debug
# Prevent supervisord from clearing any existing AUTO child log files at
# startup time. Useful for debugging.
nocleanup=true

[unix_http_server]
file=/var/run/supervisor.sock
chmod=0700

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
serverurl=unix:////var/run/supervisor.sock
username=admin

[program:crond]
priority=10
directory=/
command=/usr/sbin/cron -f
user=root
autostart=true
autorestart=true
stopsignal=QUIT

[program:rsyslog]
priority=11
directory=/
command=/etc/init.d/rsyslog start
user=root
autostart=true
autorestart=true
stopsignal=QUIT

[program:apache2]
priority=20
directory=/
command=/usr/sbin/apache2ctl -D FOREGROUND
user=www-data
autostart=true
autorestart=true
stopsignal=QUIT
EOF

cat >/tmp/proxy.image/Dockerfile <<-EOF
FROM debian:sid-slim AS Builder
LABEL maintainer "CS SI"

ENV DEBIAN_FRONTEND noninteractive

# Install Apache2
RUN apt-get update -y \
 && apt-get install -y apache2 python3-software-properties software-properties-common libapache2-modsecurity libapache2-mod-evasive logrotate
RUN add-apt-repository -y ppa:certbot/certbot \
 && apt-get update \
 && apt-get install -y python-certbot-apache
RUN a2enmod proxy \
 && a2enmod proxy_http \
 && a2enmod proxy_wstunnel \
 && a2enmod ssl \
 && a2enmod headers \
 && a2enmod rewrite

# Volume Creation
# Apache Conf
VOLUME /apache2-conf
# Certificate
VOLUME /certificate

# Create link to apache2 conf file
WORKDIR /etc/modsecurity/
# Remove conf file (they will be linked to the dockerfile volume)
RUN ln -s /apache2-conf/modsecurity.conf
WORKDIR /etc/apache2/mods-available
RUN rm -rf evasive.conf
RUN ln -s /apache2-conf/evasive.conf

# Add startup script
RUN mkdir /opt/safescale
WORKDIR /opt/safescale
ADD startup.sh .
ADD generateCertAndKeys.sh .
RUN chmod 755 /opt/safescale/*.sh

# Store default conf files in /data/docker-conf/
# This conf will update used conf file if more recent (see Scripts/startup.sh)
RUN mkdir -p /data/docker-conf/apache2-conf/
ADD ./apache2-conf/ /data/docker-conf/apache2-conf/
RUN mkdir -p /data/docker-conf/Key/
ADD ./Key/ /data/docker-conf/Key/
RUN mkdir -p /data/docker-conf/logrotate.d/
ADD ./logrotate.d/ /data/docker-conf/logrotate.d/
RUN mkdir -p /data/docker-conf/sites-available/
ADD ./sites-available/ /data/docker-conf/sites-available/

# Change group so that logrotate can run without the syslog group
RUN sed -i 's/su root syslog/su root adm/' /etc/logrotate.conf

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/opt/safescale/startup.sh"]
EOF
docker build -t proxy:latest /tmp/proxy.image

docker save proxy:latest | pigz /usr/local/dcos/genconf/serve/docker/proxy.tar.gz || exit 1
rm -rf /tmp/proxy.image
