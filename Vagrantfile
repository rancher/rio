# -*- mode: ruby -*-
# vi: set ft=ruby :

# Specify minimum Vagrant version and Vagrant API version
Vagrant.require_version '>= 1.6.0'
VAGRANTFILE_API_VERSION = '2'

# Explicitly specify only supported provider
ENV['VAGRANT_DEFAULT_PROVIDER'] = 'virtualbox'

# Read YAML config file
CONFIG = YAML.load_file(File.join(File.dirname(__FILE__), 'vagrant/config.yaml'))

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
  config.vm.box = "ubuntu/xenial64"
  
  # forward rio plaintext/tls ports to localhost
  config.vm.network "forwarded_port", guest: 7080, host: 7080, host_ip: "127.0.0.1"
  config.vm.network "forwarded_port", guest: 7443, host: 7443, host_ip: "127.0.0.1"

  # add bridge network, requires user input (select interface to bridge)
  if CONFIG['network'] == "public"
    config.vm.network "public_network"
  end
  
  config.vm.provider "virtualbox" do |vb|
    vb.cpus = CONFIG['machine']['cpu']
    vb.memory = CONFIG['machine']['memory']
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
    config.vm.provision "shell", inline: download_unix(
      CONFIG['version'], target)
  end

  # install and start rio server on guest
  config.vm.provision "shell", inline: install_unix(
    CONFIG['version'], "/vagrant/.vagrant", "/usr/bin", "linux-amd64")
  config.vm.provision "shell", inline: daemonize
  config.vm.provision "shell", inline: login_unix('guest', 'root')
  config.vm.provision "shell", inline: login_unix('guest', 'vagrant')

  # install and configure rio client on host
  if OS.windows?
    config.vm.provision "host_shell", inline: install_windows(
      CONFIG['version'], "C:\\Windows\\system32")
    config.vm.provision "host_shell", inline: login_windows
  else
    config.vm.provision "host_shell", inline: install_unix(
      CONFIG['version'], ".vagrant", "/usr/local/bin", hostOS)
    config.vm.provision "host_shell", inline: login_unix('host', '')
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
  target == 'guest' ?
    token = 'sudo cat /var/lib/rancher/rio/server/client-token 2>/dev/null' :
    token = 'vagrant ssh -c \'sudo cat /var/lib/rancher/rio/server/client-token\' 2>/dev/null'
  userset = (user != '')

  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    while true; do
      token=`#{token}`
      if [ "$token" != "" ]; then
        if #{userset}; then
          sudo -H -u #{user} bash -c \
            "rio login --server https://127.0.0.1:7443 --token $token 2>&1"
        else
          rio login --server https://127.0.0.1:7443 --token $token 2>&1
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
      $token = (vagrant ssh -c 'sudo cat /var/lib/rancher/rio/server/client-token');
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

def daemonize()
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
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
