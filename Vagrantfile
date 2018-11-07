# -*- mode: ruby -*-
# vi: set ft=ruby :

# Specify minimum Vagrant version and Vagrant API version
Vagrant.require_version '>= 1.6.0'
VAGRANTFILE_API_VERSION = '2'

# Read YAML config file
CONFIG = YAML.load_file(File.join(File.dirname(__FILE__), 'vagrant.yaml'))

ENV['VAGRANT_DEFAULT_PROVIDER'] = CONFIG['provider']

case CONFIG['provider']
when "virtualbox"
  IFACE="enp0s8"
when "vmware_fusion"
  IFACE="ens38"
else
  puts "Unsupported vagrant provider: #{CONFIG['provider']}"
  Kernel.exit(1)
end

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

if CONFIG['detect_interface']
  case CONFIG['provider']
  when "virtualbox"
    host_interfaces = %x( VBoxManage list bridgedifs | grep ^Name ).gsub(/Name:\s+/, '').split("\n")
    preferred_interfaces = ['Ethernet', 'eth0', 'en0', 'Wi-Fi', 'Thunderbolt 1', 'Thunderbolt 2']
    interface_to_use = preferred_interfaces.map{ |pi| host_interfaces.find { |vm| vm =~ /#{Regexp.quote(pi)}/ } }.compact[0]
    if interface_to_use != nil
      puts "\nBridge Interface: #{interface_to_use}\nIf this is wrong, set \"detect_interface: false\" in vagrant.yaml\n\n"
    else
      puts "\nNo Bridge Interface Detected.\nCandidate Interfaces: #{host_interfaces}\nPreferred Interfaces: #{preferred_interfaces}\n\n"
    end
  end
end

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  CONFIG['nodes'].each_with_index do |node, i|
    config.vm.define node['name'] do |named|
      named.vm.hostname = node['name']

      case CONFIG['provider']
      when "virtualbox"
        named.vm.box = "ubuntu/xenial64"
        named.vm.network "public_network", bridge: interface_to_use
        named.vm.provider CONFIG['provider'] do |provider|
          provider.linked_clone = true if Gem::Version.new(Vagrant::VERSION) >= Gem::Version.new('1.8.0')
          provider.cpus = node['cpu']
          provider.memory = node['memory']
        end
      when "vmware_fusion"
        named.vm.box = "jamesoliver/xenial64"
        named.vm.network "public_network"
        named.vm.provider CONFIG['provider'] do |provider|
          provider.vmx['numvcpus'] = node['cpu']
          provider.vmx['memsize'] = node['memory']
        end
      end

      # install server on first encountered node
      if i == 0
        # forward rio plaintext/tls ports to localhost
        named.vm.network "forwarded_port", guest: 7080, host: 7080, host_ip: "127.0.0.1"
        named.vm.network "forwarded_port", guest: 7443, host: 7443, host_ip: "127.0.0.1"

        # detect host os
        if OS.mac?
          hostOS = "darwin"
        elsif OS.linux?
          hostOS = "linux-amd64"
        else
          hostOS = "windows"
        end

        if CONFIG['version'] != "dev"
          # download requisite binaries on guest
          Set.new(["linux-amd64", hostOS]).each do |target|
            named.vm.provision "shell", inline: download_unix(
              CONFIG['version'], target)
          end
        end

        # install and start rio server on guest
        named.vm.provision "shell", inline: install_unix(
          CONFIG['version'], "/vagrant/.vagrant", "/usr/bin", "linux-amd64")
        named.vm.provision "shell", inline: daemonize_server(node['name'])
        named.vm.provision "shell", inline: login_unix('guest', 'root')
        named.vm.provision "shell", inline: login_unix('guest', 'vagrant')

        # install and configure rio client on host
        if OS.windows?
          named.vm.provision "host_shell", inline: install_windows(
            CONFIG['version'], "C:\\Windows\\system32")
          named.vm.provision "host_shell", inline: login_windows
        else
          named.vm.provision "host_shell", inline: install_unix(
            CONFIG['version'], ".vagrant", "/usr/local/bin", hostOS)
          named.vm.provision "host_shell", inline: login_unix('host', '')
        end
      else
        # install and start rio agent on guest
        named.vm.provision "shell", inline: install_unix(
          CONFIG['version'], "/vagrant/.vagrant", "/usr/bin", "linux-amd64")
        named.vm.provision "shell", inline: daemonize_agent(node['name'])
        named.vm.provision "shell", inline: login_unix('guest', 'root')
        named.vm.provision "shell", inline: login_unix('guest', 'vagrant')
      end
    end
  end
end

def download_unix(version, rioOSArch)
  rioOSArch == "windows" ? ext = "zip" : ext = "tar.gz"
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    RIO_FILENAME=rio-#{version}-#{rioOSArch}.#{ext}
    RIO_URLBASE=https://github.com/rancher/rio/releases/download
    CHECKSUM_URL=${RIO_URLBASE}/#{version}/sha256sum.txt
    CHECKSUM_FILEPATH=/vagrant/.vagrant/sha256sum.txt
    RIO_URL=${RIO_URLBASE}/#{version}/${RIO_FILENAME}
    RIO_FILEPATH=/vagrant/.vagrant/${RIO_FILENAME}

    echo Downloading Rio #{version} checksums
    curl -sSfL ${CHECKSUM_URL} -o ${CHECKSUM_FILEPATH}
    EXPECT=$(grep "dist/artifacts/${RIO_FILENAME}" ${CHECKSUM_FILEPATH} | cut -d ' ' -f 1)

    DOWNLOAD=1
    if [ -f ${RIO_FILEPATH} ]; then
      echo Checking existing copy of Rio #{version} for #{rioOSArch}
      HAVE=$(sha256sum ${RIO_FILEPATH} | cut -d ' ' -f 1)
      if [ "${HAVE}" = "${EXPECT}" ]; then
        DOWNLOAD=0
      fi
    fi

    if [ $DOWNLOAD -eq 1 ]; then
      echo Downloading Rio #{version} for #{rioOSArch}...
      curl -sSfL ${RIO_URL} -o ${RIO_FILEPATH}
      HAVE=$(sha256sum ${RIO_FILEPATH} | cut -d ' ' -f 1)
      if [ "${HAVE}" = "${EXPECT}" ]; then
        echo Downloaded Rio #{version} to ${RIO_FILEPATH}
      else
        echo "Download of ${RIO_FILEPATH} did not match expected SHA"
        exit 1
      fi
    else
      echo Rio #{version} for #{rioOSArch} already downloaded
    fi
  EOF
end

def install_unix(version, installPath, binPath, rioOSArch)
  case version
  when "dev"
    case rioOSArch
    when "linux-amd64"
      binFile = "rio"
    when "darwin"
      binFile = "rio-darwin"
    end
    return <<-EOF
      if #{CONFIG['debug']}; then set -x; fi
      RIO_FILEPATH=#{installPath}/../bin/#{binFile}

      if [ ! -f #{binPath}/rio ] || [[ `rio -v|tr ' ' '\n'|tail -n1` != `${RIO_FILEPATH} -v|tr ' ' '\n'|tail -n1` ]]; then
        if [ ! -f ${RIO_FILEPATH} ]; then
          echo "Couldn't find file: ${RIO_FILEPATH}"
          exit 1
        fi
        rm -f #{binPath}/rio
        sudo cp #{installPath}/../bin/#{binFile} #{binPath}/rio
        echo Installed Rio #{version} to #{binPath}
      else
        echo Rio #{version} already installed
      fi
    EOF
  else
    return <<-EOF
      if #{CONFIG['debug']}; then set -x; fi
      RIO_FILEPATH=#{installPath}/rio-#{version}-#{rioOSArch}.tar.gz

      if [ ! -f #{binPath}/rio ] || [[ `rio -v | grep #{version}$` == "" ]]; then
        if [ ! -f ${RIO_FILEPATH} ]; then
          echo "Couldn't find file: ${RIO_FILEPATH}"
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
      if [ #{target} == "guest" ]; then
        token=`cat /vagrant/.vagrant/client-token`
        ip=`cat /vagrant/.vagrant/node-ip`
      else
        token=`cat .vagrant/client-token`
        ip=127.0.0.1
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

def daemonize_server(name)
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    # determine ipv4 address of bridged network
    node_ip=`ifconfig #{IFACE}|grep 'inet '|tr ' ' '\n'|grep addr|tr ':' '\n'|tail -n1`

    # add rio to systemd and start it
    sudo systemctl stop rio || true
    sudo cat << F00F > /etc/systemd/system/multi-user.target.wants/rio.service
[Unit]
Description=Rio Server
After=network.target
ConditionPathExists=/usr/bin/rio

[Service]
ExecStart=/usr/bin/rio server \
--log /vagrant/.vagrant/#{name}.log \
--node-ip ${node_ip}
KillMode=process
Restart=on-failure

[Install]
WantedBy=multi-user.target
Alias=rio.service
F00F
    sudo systemctl daemon-reload
    sudo systemctl restart rio
    echo rio server started, waiting for tokens
    while true; do
      if [ -f /var/lib/rancher/rio/server/client-token ]; then
        cp /var/lib/rancher/rio/server/client-token /vagrant/.vagrant/client-token
        if [ -f /var/lib/rancher/rio/server/node-token ]; then
          cp /var/lib/rancher/rio/server/node-token /vagrant/.vagrant/node-token
          echo ${node_ip} > /vagrant/.vagrant/node-ip
          echo tokens copied successfully
          break
        fi
      fi
      sleep 1
    done
  EOF
end

def daemonize_agent(name)
  return <<-EOF
    if #{CONFIG['debug']}; then set -x; fi
    node_ip=`ifconfig #{IFACE}|grep 'inet '|tr ' ' '\n'|grep addr|tr ':' '\n'|tail -n1`
    server_ip=`cat /vagrant/.vagrant/node-ip`
    token=`cat /vagrant/.vagrant/node-token`

    # add rio to systemd and start it
    sudo systemctl stop rio || true
    sudo cat << F00F > /etc/systemd/system/multi-user.target.wants/rio.service
[Unit]
Description=Rio Agent
After=network.target
ConditionPathExists=/usr/bin/rio

[Service]
ExecStart=/usr/bin/rio agent \
--server https://${server_ip}:7443 \
--token ${token} \
--log /vagrant/.vagrant/#{name}.log \
--node-ip ${node_ip}
KillMode=process
Restart=on-failure

[Install]
WantedBy=multi-user.target
Alias=rio.service
F00F
    sudo systemctl daemon-reload
    sudo systemctl restart rio
    echo rio agent started
  EOF
end
