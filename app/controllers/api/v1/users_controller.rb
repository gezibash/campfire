module Api
  module V1
    class UsersController < BaseController
      def index
        users = User.active.ordered
        render json: users.map { |u| user_json(u) }
      end

      # PATCH /api/v1/users/:id/presence — update last_seen_at for a user.
      def presence
        user = User.active.find_by(id: params[:id])
        unless user
          render json: { error: "User not found" }, status: :not_found
          return
        end

        user.update_column(:last_seen_at, Time.current)
        render json: { ok: true, user_id: user.id, last_seen_at: user.last_seen_at.iso8601 }
      end

      # GET /api/v1/users/presence — bulk presence for all active users.
      def presence_index
        users = User.active.select(:id, :name, :last_seen_at)
        render json: users.map { |u|
          { id: u.id, name: u.name, last_seen_at: u.last_seen_at&.iso8601 }
        }
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

        user.avatar.attach(params[:avatar]) if params[:avatar].present?

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
