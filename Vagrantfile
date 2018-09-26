# -*- mode: ruby -*-
# vi: set ft=ruby :

# Specify minimum Vagrant version and Vagrant API version
Vagrant.require_version '>= 1.6.0'
VAGRANTFILE_API_VERSION = '2'

# Read YAML config file
CONFIG = YAML.load_file(File.join(File.dirname(__FILE__), 'vagrant.yaml'))

ENV['VAGRANT_DEFAULT_PROVIDER'] = CONFIG['provider']

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

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  case CONFIG['provider']
  when "virtualbox"
    config.vm.box = "ubuntu/xenial64"
    config.vm.provider "virtualbox" do |provider|
      provider.linked_clone = true if Gem::Version.new(Vagrant::VERSION) >= Gem::Version.new('1.8.0')
      provider.cpus = CONFIG['machine']['cpu']
      provider.memory = CONFIG['machine']['memory']
    end
  when "vmware_fusion"
    config.vm.box = "jamesoliver/xenial64"
    config.vm.provider "vmware_fusion" do |provider|
      provider.linked_clone = true if Gem::Version.new(Vagrant::VERSION) >= Gem::Version.new('1.8.0')
      provider.vmx['memsize'] = CONFIG['machine']['memory']
      provider.vmx['numvcpus'] = CONFIG['machine']['cpu']
    end
  else
    puts "Unsupported vagrant provider: #{CONFIG['provider']}"
    Kernel.exit(1)
  end

  config.vm.define "node-01" do |server|
    server.vm.hostname = "node-01"

    # forward rio plaintext/tls ports to localhost
    server.vm.network "forwarded_port", guest: 7080, host: 7080, host_ip: "127.0.0.1"
    server.vm.network "forwarded_port", guest: 7443, host: 7443, host_ip: "127.0.0.1"

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
      server.vm.provision "shell", inline: download_unix(
        CONFIG['version'], target)
    end

    # install and start rio server on guest
    server.vm.provision "shell", inline: install_unix(
      CONFIG['version'], "/vagrant/.vagrant", "/usr/bin", "linux-amd64")
    server.vm.provision "shell", inline: default_route_hack
    server.vm.provision "shell", inline: daemonize_server
    server.vm.provision "shell", inline: login_unix('guest', 'root')
    server.vm.provision "shell", inline: login_unix('guest', 'vagrant')

    # install and configure rio client on host
    if OS.windows?
      server.vm.provision "host_shell", inline: install_windows(
        CONFIG['version'], "C:\\Windows\\system32")
      server.vm.provision "host_shell", inline: login_windows
    else
      server.vm.provision "host_shell", inline: install_unix(
        CONFIG['version'], ".vagrant", "/usr/local/bin", hostOS)
      server.vm.provision "host_shell", inline: login_unix('host', '')
    end
  end

  for i in 2..CONFIG['nodes']
    name = "node-%02d" %[i]
    config.vm.define name do |agent|
      agent.vm.hostname = name

      # install and start rio agent on guest
      agent.vm.provision "shell", inline: install_unix(
        CONFIG['version'], "/vagrant/.vagrant", "/usr/bin", "linux-amd64")
      agent.vm.provision "shell", inline: default_route_hack
      agent.vm.provision "shell", inline: daemonize_agent(name)
    end
  end

  # add bridge network, requires user input (select interface to bridge)
  if CONFIG['network'] == "public"
    config.vm.network "public_network"
  end
end

def download_unix(version, rioOSArch)
  rioOSArch == "windows" ? ext = "zip" : ext = "tar.gz"
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    RIO_FILENAME=rio-#{version}-#{rioOSArch}.#{ext}
    RIO_URLBASE=https://github.com/rancher/rio/releases/download
    RIO_URL=${RIO_URLBASE}/#{version}/${RIO_FILENAME}
    RIO_FILEPATH=/vagrant/.vagrant/${RIO_FILENAME}

    if [ ! -f ${RIO_FILEPATH} ]; then
      echo Downloading Rio #{version} for #{rioOSArch}...
      curl -sSfL ${RIO_URL} -o ${RIO_FILEPATH}
      echo Downloaded Rio #{version} to ${RIO_FILEPATH}
    else
      echo Rio #{version} for #{rioOSArch} already downloaded
    fi
  EOF
end

def install_unix(version, installPath, binPath, rioOSArch)
  rioOSArch == "windows" ? ext = "zip" : ext = "tar.gz"
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    RIO_FILENAME=rio-#{version}-#{rioOSArch}.#{ext}
    RIO_FILEPATH=#{installPath}/${RIO_FILENAME}

    if [ ! -f #{binPath}/rio ] || [[ `rio -v | grep #{version}` == "" ]]; then
      if [ ! -f ${RIO_FILEPATH} ]; then
        echo "Couldn't find file: ${RIO_FILENAME}"
        exit 1
      fi
      rm -f #{binPath}/rio
      tar xvzf ${RIO_FILEPATH} -C #{binPath} --strip-components 1
      echo Installed Rio #{version} to #{binPath}
    else
      echo Rio #{version} already installed
    fi
  EOF
end

def install_windows(version, binPath)
  return <<-EOF
    if ($#{CONFIG['debug']}) { Set-PSDebug -Trace 1 }
    if ((Test-Path #{binPath}\\rio.exe) -And ((rio.exe -v) -like ('*#{version}'))) {
      Write-Host Rio #{version} already installed;
    } else {
      $tempdir = [System.IO.Path]::GetTempPath();
      $dirname = 'rio-#{version}-windows';
      $filename = $dirname + '.zip';
      Add-Type -A System.IO.Compression.FileSystem;
      [IO.Compression.ZipFile]::ExtractToDirectory('.vagrant\\' + $filename, $tempdir);
      Start-Process -FilePath powershell.exe -Verb runAs -Wait -ArgumentList \\"-Command Move-Item -Path $tempdir$dirname\\rio.exe -Destination #{binPath}\\rio.exe -Force\\";
      # process completes before file is moved...
      Start-Sleep -Seconds 1;
      Remove-Item $tempdir$dirname -Recurse -Force;
      if (Test-Path #{binPath}\\rio.exe) {
        Write-Host Installed Rio #{version} to #{binPath};
      } else {
        Write-Host Failed to install Rio #{version} to #{binPath};
      }
    }
  EOF
end

def login_unix(target, user)
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    while true; do
      ip=127.0.0.1
      if [ #{target} == "guest" ]; then
        token=`cat /vagrant/.vagrant/client-token`
      else
        token=`cat .vagrant/client-token`
        if [ "#{CONFIG['provider']}" == "vmware_fusion" ]; then
          ip=`vagrant ssh-config node-01 | grep HostName | sed 's/^.*HostName //g'`
        fi
      fi
      if [ "$token" != "" ]; then
        if [ "#{user}" != "" ]; then
          sudo -H -u #{user} bash -c \
            "rio login --server https://${ip}:7443 --token $token 2>&1"
        else
          rio login --server https://${ip}:7443 --token $token 2>&1
        fi
        if [ $? -eq 0 ]; then
          break
        fi
      fi
      sleep 1
    done
  EOF
end

def login_windows()
  return <<-EOF
    if ($#{CONFIG['debug']}) { Set-PSDebug -Trace 1; }
    while($true) {
      $token = (vagrant ssh node-01 -c 'sudo cat /var/lib/rancher/rio/server/client-token');
      if ($token -ne '') {
        $result = rio login --server https://127.0.0.1:7443 --token $token 2>&1
        if ($result -like ('*Log in successful*')) {
          Write-Host $result
          break
        }
      }
      Start-Sleep -Seconds 1;
    }
  EOF
end

# configure default route so rio chooses the correct IP address
# TODO: use rio config param once available
def default_route_hack()
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    route del default gw 10.0.2.2
    # FIXME: won't work unless this is the correct gateway for bridged network
    route add default gw 192.168.0.1
  EOF
end

def daemonize_server()
  case CONFIG['provider']
  when "virtualbox"
    iface="enp0s8"
  when "vmware_fusion"
    iface="ens38"
  else
    puts "Unsupported vagrant provider: #{CONFIG['provider']}"
    Kernel.exit(1)
  end

  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    # add rio to systemd and start it
    sudo cat << F00F > /etc/systemd/system/multi-user.target.wants/rio-server.service
[Unit]
Description=Rio Server
After=network.target
ConditionPathExists=/usr/bin/rio

[Service]
ExecStart=/usr/bin/rio server \
  --log /vagrant/.vagrant/node-01.log
KillMode=process
Restart=on-failure

[Install]
WantedBy=multi-user.target
Alias=rio.service
F00F
    sudo systemctl daemon-reload
    sudo systemctl restart rio-server
    echo rio-server started, waiting for tokens
    while true; do
      if [ -f /var/lib/rancher/rio/server/client-token ]; then
        cp /var/lib/rancher/rio/server/client-token /vagrant/.vagrant/client-token
        if [ -f /var/lib/rancher/rio/server/node-token ]; then
          cp /var/lib/rancher/rio/server/node-token /vagrant/.vagrant/node-token
          ifconfig #{iface}|grep 'inet '|tr ' ' '\n'|grep addr|tr ':' '\n'|tail -n1>/vagrant/.vagrant/node-ip
          break
        fi
      fi
      sleep 1
    done
    echo copied tokens
  EOF
end

def daemonize_agent(name)
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    # add rio to systemd and start it
    ip=`cat /vagrant/.vagrant/node-ip`
    token=`cat /vagrant/.vagrant/node-token`
    sudo cat << F00F > /etc/systemd/system/multi-user.target.wants/rio-agent.service
[Unit]
Description=Rio Agent
After=network.target
ConditionPathExists=/usr/bin/rio

[Service]
ExecStart=/usr/bin/rio agent \
  --server https://${ip}:7443 \
  --token ${token} \
  --log /vagrant/.vagrant/#{name}.log
KillMode=process
Restart=on-failure

[Install]
WantedBy=multi-user.target
Alias=rio.service
F00F
    sudo systemctl daemon-reload
    sudo systemctl restart rio-agent
    echo rio-agent started
  EOF
end
