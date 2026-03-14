module Api
  module V1
    class SessionsController < ActionController::Base
      skip_forgery_protection

      def create
        user = User.active.find_by(email_address: params[:email_address])

        if user&.authenticate(params[:password])
          user.update!(bot_token: User.generate_bot_token) unless user.bot_token.present?

          render json: {
            user: { id: user.id, name: user.name, email_address: user.email_address, role: user.role },
            api_token: user.bot_key
          }
        else
          render json: { error: "Invalid email or password" }, status: :unauthorized
        end
      end
    end
  end
end
