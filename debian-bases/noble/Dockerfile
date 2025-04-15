FROM ubuntu:22.04

MAINTAINER Herpiko Dwi Aguno <herpiko@aguno.xyz>

ENV distro noble

RUN apt-get update
RUN apt-get -y install pbuilder

COPY base.tgz /var/cache/pbuilder/base.tgz
