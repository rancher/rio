begin
  require 'vagrant'
rescue LoadError
  raise 'The vagrant-host-shell plugin must be run within Vagrant.'
end

module VagrantPlugins::HostShell
  class Plugin < Vagrant.plugin('2')
    name 'vagrant-host-shell'
    description <<-DESC.gsub(/^ +/, '')
      A simple provisioner to run commands on the host at machine
      boot, instead of the guest.
    DESC

    config(:host_shell, :provisioner) do
      require_relative 'config'
      Config
    end

    provisioner(:host_shell) do
      require_relative 'provisioner'
      Provisioner
    end
  end
end
