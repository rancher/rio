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

      if Vagrant::Util::Platform.windows?
        shell = "powershell"
        flag = "-Command"
        pathEnv = "PATH"
      else
        shell = "bash"
        flag = "-c"
        pathEnv = "VAGRANT_OLD_ENV_PATH"
      end

      result = Vagrant::Util::Subprocess.execute(
        shell,
        flag,
        config.inline,
        :notify => [:stdout, :stderr],
        :workdir => config.cwd,
        :env => {PATH: ENV[pathEnv]},
      ) do |type, data|
        handle_comm(type, data)
      end

      if config.abort_on_nonzero && !result.exit_code.zero?      
        raise VagrantPlugins::HostShell::Errors::NonZeroStatusError.new(config.inline, result.exit_code)  
      end
    end
  end
end
