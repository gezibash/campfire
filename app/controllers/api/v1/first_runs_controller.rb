module Api
  module V1
    class FirstRunsController < ActionController::Base
      skip_forgery_protection

      def create
        if Account.any?
          render json: { error: "Already set up" }, status: :unprocessable_entity
          return
        end

        user = FirstRun.create!(user_params)
        user.update!(bot_token: User.generate_bot_token) unless user.bot_token.present?

        render json: {
          user: user_json(user),
          api_token: user.bot_key
        }, status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      private
        def user_params
          params.permit(:name, :email_address, :password)
        end

        def user_json(user)
          { id: user.id, name: user.name, email_address: user.email_address, role: user.role, admin: user.administrator? }
        end
    end
  end
end
