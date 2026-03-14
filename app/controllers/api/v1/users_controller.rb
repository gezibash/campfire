module Api
  module V1
    class UsersController < BaseController
      def index
        users = User.active.ordered
        render json: users.map { |u| user_json(u) }
      end

      def create
        require_administrator!
        return if performed?

        password = params[:password].presence || SecureRandom.alphanumeric(16)
        user = User.create!(
          name: params[:name],
          email_address: params[:email_address],
          password: password,
          role: params[:role] || :member,
          bot_token: User.generate_bot_token
        )

        render json: user_json(user).merge(password: password, api_token: user.bot_key), status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      private
        def user_json(user)
          {
            id: user.id,
            name: user.name,
            email_address: user.email_address,
            role: user.role,
            admin: user.administrator?,
            api_token: user.bot_key,
            avatar_url: user.avatar.attached? ? url_for(user.avatar) : nil
          }
        end
    end
  end
end
