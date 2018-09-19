# -*- mode: ruby -*-
# vi: set ft=ruby :

DEBUG = false
RIO_VERSION = "v0.0.3-rc6"

$LOAD_PATH.unshift File.expand_path('../vagrant/lib', __FILE__)
require 'vagrant-host-shell'

def downloadRio(binPath, rioOSArch)
  return <<-EOF
    if #{DEBUG}; then set -x; fi
    RIO_BASEPATH=https://github.com/rancher/rio/releases/download
    RIO_PATH=${RIO_BASEPATH}/#{RIO_VERSION}/rio-#{RIO_VERSION}-#{rioOSArch}.tar.gz

    if [ ! -f #{binPath}/rio ] || [[ `rio -v | grep #{RIO_VERSION}` == "" ]]; then
      echo Downloading rio... this could take a while
      curl -sSfL ${RIO_PATH} | tar xvzf - -C #{binPath} --strip-components 1
      echo Download complete
    fi
    echo Done
  EOF
end

Vagrant.configure("2") do |config|
  # TODO: select custom box based on rio version
  config.vm.box = "ubuntu/trusty64"
  
  config.vm.network "forwarded_port", guest: 7080, host: 7080, host_ip: "127.0.0.1"
  config.vm.network "forwarded_port", guest: 7443, host: 7443, host_ip: "127.0.0.1"
  
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "1024"
  end

  config.vm.provision "shell", inline: downloadRio("/usr/bin", "linux-amd64")
  # TODO: detect host os (darwin and linux-amd64 are supported)
  config.vm.provision "host_shell", inline: downloadRio("/usr/local/bin", "darwin")
end
