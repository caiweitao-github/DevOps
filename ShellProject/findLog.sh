#!/bin/bash
LogFile="/data/log/jip/jip.log"
TmepFile=$(mktemp -t XXXXX)
TmepFile2=$(mktemp -t XXXXX)
cat $LogFile |grep 'lost'|awk '{print $5,$10}'|awk -F ',' '{print $1}'|grep -Ev '[2-8]m'|grep -Ev 'lost|left' > $TmepFile
cat $TmepFile |awk '{print $1}'|sort -rn|uniq -c|sort -rn|awk '{if($1 > 1) print $0}' > $TmepFile2

while read line;do
    id=$(echo $line|awk '{print $2}')
    count=$(echo $line|awk '{print $1}')
    echo "ID: $id, Count: $count"
    cat $LogFile |grep 'lost'|awk '{print $1,$2,$5,$10}'|awk -F ',' '{print $1}'|grep -Ev '[2-8]m'|grep -Ev 'lost|left'|grep "$id"
    echo "-----------------------------------------------------------------------------------------------------------------------"
done < $TmepFile2

rm -f $TmepFile $TmepFile2