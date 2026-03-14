module Api
  module V1
    class BaseController < ActionController::Base
      skip_forgery_protection

      before_action :authenticate_api_user!

      private
        def authenticate_api_user!
          token = request.headers["Authorization"]&.delete_prefix("Bearer ")&.strip
          if token.blank?
            render json: { error: "Missing Authorization header" }, status: :unauthorized
            return
          end

          user_id, bot_token = token.split("-", 2)
          @current_api_user = User.active.find_by(id: user_id, bot_token: bot_token)

          unless @current_api_user
            render json: { error: "Invalid API token" }, status: :unauthorized
            return
          end

          Current.user = @current_api_user
        end

        def require_administrator!
          unless @current_api_user&.administrator?
            render json: { error: "Administrator access required" }, status: :forbidden
          end
        end
    end
  end
end
