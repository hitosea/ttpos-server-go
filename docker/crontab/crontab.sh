#!/bin/sh
# service cron status
# * * * * * bash /var/www/docker/crontab/crontab.sh
# PATH=/usr/local/php/bin:/opt/someApp/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

# php think crontab >> /dev/null 2>&1

# step=1
# for ((i=0;i<55;i=(i+step))); do
#     sleep $step
#     curl -sS --connect-timeout 10 -m 60 '10.32.33.4/system/crontab' >> /dev/null 2>&1 &
#     # cd /var/www
# done
exit 0

