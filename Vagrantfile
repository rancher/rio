# -*- mode: ruby -*-
# vi: set ft=ruby :

DEBUG = false
RIO_VERSION = "v0.0.3-rc6"
MEMORY_MIB = "1024"

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
  config.vm.box = "ubuntu/xenial64"
  
  config.vm.network "forwarded_port", guest: 7080, host: 7080, host_ip: "127.0.0.1"
  config.vm.network "forwarded_port", guest: 7443, host: 7443, host_ip: "127.0.0.1"
  
  config.vm.provider "virtualbox" do |vb|
    vb.memory = MEMORY_MIB
  end

  config.vm.provision "shell", inline: downloadRio("/usr/bin", "linux-amd64")
  config.vm.provision "shell", inline: <<-EOF
    if #{DEBUG}; then set -x; fi
    sudo cp /vagrant/vagrant/rio.service /etc/systemd/system/multi-user.target.wants/
    sudo systemctl daemon-reload
    sudo systemctl start rio
  EOF

  # TODO: detect host os (darwin and linux-amd64 are supported)
  config.vm.provision "host_shell", inline: downloadRio("/usr/local/bin", "darwin")
  config.vm.provision "host_shell", inline: <<-EOF
    if #{DEBUG}; then set -x; fi
    while true; do
      token=`vagrant ssh -c 'sudo cat /var/lib/rancher/rio/server/client-token' 2>/dev/null`
      if [ "$token" != "" ]; then
        rio login --server https://127.0.0.1:7443 --token $token
        if [ $? -eq 0 ]; then
          break
        fi
      fi
      sleep 3
    done
  EOF
end
