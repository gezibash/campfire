class CliAuthController < ApplicationController
  def authorize
    port = params.require(:port)

    # Ensure user has a bot_token for API access
    Current.user.update!(bot_token: User.generate_bot_token) if Current.user.bot_token.blank?

    callback = "http://localhost:#{port}/callback?" + {
      token: Current.user.bot_key,
      user_id: Current.user.id,
      name: Current.user.name
    }.to_query

    redirect_to callback, allow_other_host: true
  end
end
