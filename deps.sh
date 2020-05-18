#!/bin/bash

DEPS=( $(ldd ./goyammer | grep -P -o "(?<=> ).+(?= \(0x)" | xargs dpkg -S | awk -F ':' '{print $1}' | sort | uniq) )

for element in "${DEPS[@]}"
do
    VERSION=$(aptitude show $element |grep Version | gawk '{print $2}' | grep -Po '^(?:([0-9]+:))?[0-9]+\.[0-9]+')
    echo "$element (>= ${VERSION})"
done
  
