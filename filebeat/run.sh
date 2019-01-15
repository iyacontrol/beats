#!/bin/sh
#df | grep rbd | awk '{print $1}' | xargs umount
./filebeat --c /etc/filebeat.yml --path.data /usr/share/filebeat/data --path.logs /var/log/filebeat