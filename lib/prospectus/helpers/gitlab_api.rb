Prospectus.extra_dep('gitlab_api', 'keylime')
Prospectus.extra_dep('gitlab_api', 'gitlab')

module LogCabin
  module Modules
    ##
    # Provide an api method for modules to query GitLab
    module GitlabApi
      def gitlab_api
        @gitlab_api ||= Gitlab.client(
          endpoint: gitlab_endpoint + '/api/v4',
          private_token: gitlab_token
        )
      end

      private

      def gitlab_token
        @gitlab_token ||= token_from_file
        @gitlab_token ||= Keylime.new(
          server: gitlab_endpoint,
          account: 'prospectus'
        ).get!("GitLab API token (#{gitlab_endpoint}/profile/account)").password
      end

      def token_from_file
        return unless File.exist? File.expand_path('~/.gitlab_api')
        File.read('~/.gitlab_api').strip
      end

      def gitlab_endpoint
        @gitlab_endpoint ||= 'https://gitlab.com'
      end

      def repo(value)
        @repo = value
      end

      def endpoint(value)
        @gitlab_endpoint = value
      end
    end
  end
end
