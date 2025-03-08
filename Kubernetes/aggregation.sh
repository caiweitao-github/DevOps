#!/bin/bash

cfgname=$1
fname=${cfgname}-config.yaml
k8snamespace=kdl

echo "apiVersion: v1" > ${fname}
echo "data:" >> ${fname}

DIRECTORY="./"  

find "$DIRECTORY" -type f -name "*.json" | while read -r FILE; do  
    fn=$(basename "$FILE")
    echo "  ${fn}: |-" >> ${fname}  
    sed 's/^/    /' "$FILE" >> ${fname}
    echo -e "\n" >> ${fname}
done

echo "" >> ${fname}
echo "kind: ConfigMap" >> ${fname}
echo "metadata:" >> ${fname}
echo "  name: ${cfgname}" >> ${fname}
echo "  namespace: ${k8snamespace}" >> ${fname}