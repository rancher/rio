# https://github.com/LLParse/vagrant-host-shell
module VagrantPlugins
  module HostShell
  	lib_path = Pathname.new(File.expand_path("../vagrant-host-shell", __FILE__))
  	autoload :Errors, lib_path.join("errors")
  end
end

require 'vagrant-host-shell/plugin'
