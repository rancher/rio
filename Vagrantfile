# -*- mode: ruby -*-
# vi: set ft=ruby :

DEBUG = false
RIO_VERSION = "v0.0.3-rc7"
MEMORY_MIB = "1024"
NETWORK = "public"  # public or nat, rio services aren't accessible behind nat

require 'set'
$LOAD_PATH.unshift File.expand_path('../vagrant/lib', __FILE__)
require 'vagrant-host-shell'

module OS
  def OS.windows?
        (/cygwin|mswin|mingw|bccwin|wince|emx/ =~ RUBY_PLATFORM) != nil
  end
  def OS.mac?
    (/darwin/ =~ RUBY_PLATFORM) != nil
  end
  def OS.unix?
    !OS.windows?
  end
  def OS.linux?
    OS.unix? and not OS.mac?
  end
end

Vagrant.configure("2") do |config|
  # TODO: select custom box based on rio version
  config.vm.box = "ubuntu/xenial64"
  
  # forward rio plaintext/tls ports to localhost
  config.vm.network "forwarded_port", guest: 7080, host: 7080, host_ip: "127.0.0.1"
  config.vm.network "forwarded_port", guest: 7443, host: 7443, host_ip: "127.0.0.1"

  # add bridge network, requires user input (select interface to bridge)
  if NETWORK == "public"
    config.vm.network "public_network"
  end
  
  config.vm.provider "virtualbox" do |vb|
    vb.memory = MEMORY_MIB
  end

  # detect host os
  if OS.mac?
    hostOS = "darwin"
  elsif OS.linux?
    hostOS = "linux-amd64"
  else
    hostOS = "windows"
  end

  # download requisite binaries in guest
  Set.new(["linux-amd64", hostOS]).each do |target|
    config.vm.provision "shell", inline: downloadRio(target)
  end

  # install and start rio server on guest
  config.vm.provision "shell", inline: installRio("/vagrant/.vagrant", "/usr/bin", "linux-amd64")
  config.vm.provision "shell", inline: daemonize
  config.vm.provision "shell", inline: guestLogin

  # install and configure rio client on host
  if OS.windows?
    puts 'WARNING: windows host provisioning not implemented'  # TODO
  else
    config.vm.provision "host_shell", inline: installRio(".vagrant", "/usr/local/bin", hostOS)
    config.vm.provision "host_shell", inline: hostLogin
  end
end

def downloadRio(rioOSArch)
  return <<-EOF
    if #{DEBUG}; then set -x; fi
    RIO_FILENAME=rio-#{RIO_VERSION}-#{rioOSArch}.tar.gz
    RIO_URLBASE=https://github.com/rancher/rio/releases/download
    RIO_URL=${RIO_URLBASE}/#{RIO_VERSION}/${RIO_FILENAME}
    RIO_FILEPATH=/vagrant/.vagrant/${RIO_FILENAME}

    if [ ! -f ${RIO_FILEPATH} ]; then
      echo Downloading Rio #{RIO_VERSION} for #{rioOSArch}...
      curl -sSfL ${RIO_URL} -o ${RIO_FILEPATH}
      echo Downloaded Rio #{RIO_VERSION} to ${RIO_FILEPATH}
    else
      echo Rio #{RIO_VERSION} for #{rioOSArch} already downloaded
    fi
  EOF
end

def installRio(installPath, binPath, rioOSArch)
  return <<-EOF
    if #{DEBUG}; then set -x; fi
    RIO_FILENAME=rio-#{RIO_VERSION}-#{rioOSArch}.tar.gz
    RIO_FILEPATH=#{installPath}/${RIO_FILENAME}

    if [ ! -f #{binPath}/rio ] || [[ `rio -v | grep #{RIO_VERSION}` == "" ]]; then
      if [ ! -f ${RIO_FILEPATH} ]; then
        echo "Couldn't find file: ${RIO_FILENAME}"
        exit 1
      fi
      rm -f #{binPath}/rio
      tar xvzf ${RIO_FILEPATH} -C #{binPath} --strip-components 1
      echo Installed Rio #{RIO_VERSION} to #{binPath}
    else
      echo Rio #{RIO_VERSION} already installed
    fi
  EOF
end

def guestLogin()
  return <<-EOF
    if #{DEBUG}; then set -x; fi
    while true; do
      token=`sudo cat /var/lib/rancher/rio/server/client-token 2>/dev/null`
      if [ "$token" != "" ]; then
        rio login --server https://127.0.0.1:7443 --token $token
        if [ $? -eq 0 ]; then
          break
        fi
      fi
      sleep 1
    done
  EOF
end

def hostLogin()
  return <<-EOF
    if #{DEBUG}; then set -x; fi
    while true; do
      token=`vagrant ssh -c 'sudo cat /var/lib/rancher/rio/server/client-token' 2>/dev/null`
      if [ "$token" != "" ]; then
        rio login --server https://127.0.0.1:7443 --token $token
        if [ $? -eq 0 ]; then
          break
        fi
      fi
      sleep 1
    done
  EOF
end

def daemonize()
  return <<-EOF
    if #{DEBUG}; then set -x; fi
    # TODO: detect gateway? override cloud-init? rio config param?
    # hack: configure default route so rio chooses the correct IP address
    route del default gw 10.0.2.2
    route add default gw 192.168.0.1
    # add rio to systemd and start it
    sudo cp /vagrant/vagrant/rio.service /etc/systemd/system/multi-user.target.wants/
    sudo systemctl daemon-reload
    sudo systemctl restart rio
    echo Rio started
  EOF
end
