#!/bin/sh
dpkg-source --build .
pbuilder --build ../*.dsc
mkdir -p /result
cp -vR /var/cache/pbuilder/result/* /result/
