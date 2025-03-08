#!/bin/bash
LogFile=${1:-"/data/log/jip/jip.log"}
Hour=$(date +%H)
now=$(date +"%F")
TmepFile=$(mktemp -t XXXXX)
TmepFile2=$(mktemp -t XXXXX)

for i in $(seq $Hour);do
    if [[ $i -lt 10 ]];then
        i="0$i"
    fi
    cat $LogFile |grep "$now $i:"|grep 'lost'|awk '{print $5,$10}'|awk -F ',' '{print $1}'|grep -Ev 'lost|left'|awk -F '[ s]' '{if($2 < 120) print $0}' > $TmepFile
    echo "$now $i:00:00"
    cat $TmepFile |awk '{print $1}'|sort -rn|uniq -c|sort -rn|awk '{if($1 > 1) print $0}'
    echo "-----------------------------------------------------------------------------------------------------------------------"
done

rm -f $TmepFile $TmepFile2