#!/usr/bin/env python2.7
#encoding=utf-8

from fabric.api import *
from fabric.contrib.files import exists, contains, append

#from fabric.utils import abort
#from fabric.decorators import hosts, roles

import sys
sys.dont_write_bytecode = True


env.use_ssh_config = True
env.ssh_config_path = './ssh.config'
env.key_filename = '~/.ssh/google_compute_engine'
env.forward_agent = True

# host configure

hosts_web = ['35.203.148.19']
hosts_uat  = ['35.203.148.19']
hosts_all = hosts_web

# my default server
env.hosts = hosts_web

def uat():
    """
    Run on uat host
    """
    env.hosts = hosts_uat

def all():
    """
    Run on api, porter, provision, worker host
    """
    env.hosts = hosts_all

env.roledefs.update({
    'uat': hosts_uat,
    'prd': hosts_all,
})

def build(target=""):
    """
    Build all linux/amd64 binary then put to /tmp folder
    """
    with prefix('export GOOS=linux GOARCH=amd64'):
        if target == "":
            local('cd ../crawler && go build -o /tmp/crawler')
            local('upx /tmp/crawler')
            local('cd ../api && esc -o static.go static && go build -o /tmp/api')
            local('upx /tmp/api')
        elif target == "crawler":
            local('cd ../crawler && go build -o /tmp/crawler')
            local('upx /tmp/crawler')
        elif target == "api":
            local('cd ../api && esc -o static.go static && go build -o /tmp/api')
            local('upx /tmp/api')

def upload(target=""):
    """
    Upload all linux/amd64 binary from local build /tmp to GCE
    """
    # precheck for all folder
    if not exists("/usr/src/app/crawler"):
        sudo('mkdir -p /usr/src/app/crawler')
    if not exists("/usr/src/app/api/certs"):
        sudo('mkdir -p /usr/src/app/api/certs')

    if target == "":
        put('/tmp/crawler', '/usr/src/app/crawler/goapp', mode="0755", use_sudo=True)
        put('/tmp/api', '/usr/src/app/api/goapp', mode="0755", use_sudo=True)
    elif target == "crawler":
        put('/tmp/crawler', '/usr/src/app/crawler/goapp', mode="0755", use_sudo=True)
    elif target == "api":
        put('/tmp/api', '/usr/src/app/api/goapp', mode="0755", use_sudo=True)

def uptime():
    run('uptime')

def docker(cmd):
    """
    Run docker cmd
    """
    sudo("docker %s" % cmd)

def run(cmd):
    """
    Run sudo command
    """
    sudo(cmd)


def buildimage():
    if exists("/tmp/Dockerfile"):
        sudo("rm /tmp/Dockerfile")

    put('Dockerfile', '/tmp/Dockerfile', mode="0644")
    sudo('cd /tmp/ && docker build -t goapp .')

def deploy(target=""):
    build(target)
    upload(target)

def cool(target=""):
    build(target)
    upload(target)
    if target == "":
        sudo("docker restart crawler" )
    elif target == "crawler":
        sudo("docker restart %s" % target)
    elif target == "api":
        sudo("docker restart %s" % target)
