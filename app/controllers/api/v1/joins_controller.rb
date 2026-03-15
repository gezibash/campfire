module Api
  module V1
    class JoinsController < ActionController::Base
      skip_forgery_protection

      def create
        unless Current.account.join_code == params[:join_code]
          render json: { error: "Invalid join code" }, status: :unprocessable_entity
          return
        end

        user = User.create!(
          name: params[:name],
          email_address: params[:email_address],
          password: params[:password]
        )
        user.update!(bot_token: User.generate_bot_token) unless user.bot_token.present?

        render json: {
          user: { id: user.id, name: user.name, email_address: user.email_address, role: user.role },
          api_token: user.bot_key
        }, status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      rescue ActiveRecord::RecordNotUnique
        render json: { error: "Email address already taken" }, status: :unprocessable_entity
      end
    end
  end
end
