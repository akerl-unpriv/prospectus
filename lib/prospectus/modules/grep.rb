module LogCabin
  module Modules
    ##
    # Pull state from a local file
    module Grep
      include Prospectus.helpers.find(:regex)

      def load!
        raise('No file specified') unless @file
        @find ||= '.*'
        line = read_file
        @state.value = regex_helper(line)
      end

      private

      def read_file
        File.read(@file).each_line do |line|
          line = line.chomp
          return line if line.match(@find)
        end
        raise("No lines in #{@file} matched #{@find}")
      end

      def file(value)
        @file = value
      end
    end
  end
end
