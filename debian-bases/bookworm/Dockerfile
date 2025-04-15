FROM debian:latest

MAINTAINER Herpiko Dwi Aguno <herpiko@aguno.xyz>

ENV distro bookworm

RUN apt-get update
RUN apt-get -y install pbuilder

COPY base.tgz /var/cache/pbuilder/base.tgz
