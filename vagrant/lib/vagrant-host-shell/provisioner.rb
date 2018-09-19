module VagrantPlugins::HostShell
  class Provisioner < Vagrant.plugin('2', :provisioner)

    # This handles outputting the communication data back to the UI
    def handle_comm(type, data)
      if [:stderr, :stdout].include?(type)
        # Output the data with the proper color based on the stream.
        color = type == :stdout ? :green : :red

        # Clear out the newline since we add one
        data = data.chomp
        return if data.empty?

        options = {}
        options[:color] = color

        @machine.ui.detail(data.chomp, options)
      end
    end

    def provision
      @machine.ui.detail(I18n.t("vagrant.provisioners.shell.running",
        script: "host inline script"))

      result = Vagrant::Util::Subprocess.execute(
        'bash',
        '-c',
        config.inline,
        :notify => [:stdout, :stderr],
        :workdir => config.cwd,
        :env => {PATH: ENV["VAGRANT_OLD_ENV_PATH"]},
      ) do |type, data|
        handle_comm(type, data)
      end

      if config.abort_on_nonzero && !result.exit_code.zero?      
        raise VagrantPlugins::HostShell::Errors::NonZeroStatusError.new(config.inline, result.exit_code)  
      end

    end
  end
end
