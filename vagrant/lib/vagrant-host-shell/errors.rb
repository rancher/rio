module VagrantPlugins::HostShell::Errors
  
    class VagrantHostShellError < Vagrant::Errors::VagrantError; end
    
    class NonZeroStatusError < VagrantHostShellError
      def initialize(command, exit_code)        
        @command = command
        @exit_code = exit_code
        super nil
      end

      def error_message 
        "Command [#{@command}] exited with non-zero status [#{@exit_code}]"
      end

    end
end
