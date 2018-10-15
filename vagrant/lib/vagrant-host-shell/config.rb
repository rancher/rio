module VagrantPlugins::HostShell
  class Config < Vagrant.plugin('2', :config)
    attr_accessor :inline
    attr_accessor :cwd
    attr_accessor :abort_on_nonzero

    def initialize
      @inline = UNSET_VALUE
      @cwd = UNSET_VALUE
      @abort_on_nonzero = UNSET_VALUE
    end

    def finalize!
      @inline = nil if @inline == UNSET_VALUE
      @cwd = nil if @cwd == UNSET_VALUE
      @abort_on_nonzero = false if @abort_on_nonzero == UNSET_VALUE
    end

    def validate(machine)
      errors = _detected_errors

      unless inline
        errors << ':host_shell provisioner requires inline to be set'
      end

      unless abort_on_nonzero.is_a?(TrueClass) || abort_on_nonzero.is_a?(FalseClass)
        errors << ':host_shell provisioner requires abort_on_nonzero to be a boolean'
      end
      
      unless cwd.is_a?(String) || cwd.nil?
        errors << ':host_shell provisioner requires cwd to be a string or nil'
      end

      { 'host shell provisioner' => errors }
    end
  end
end
