module Api
  module V1
    class MessagesController < BaseController
      include ActiveStorage::SetCurrent

      def index
        room = @current_api_user.rooms.find(params[:room_id])
        messages = room.messages.with_creator.ordered

        if params[:after].present?
          messages = messages.where("id > ?", params[:after])
        end
        if params[:before].present?
          messages = messages.where("id < ?", params[:before])
        end

        messages = messages.limit(params.fetch(:limit, 50).to_i.clamp(1, 200))

        render json: messages.map { |m| message_json(m) }
      end

      def create
        room = @current_api_user.rooms.find(params[:room_id])

        Current.user = @current_api_user
        message = room.messages.create!(body: params[:body])
        message.broadcast_create

        render json: message_json(message), status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      private
        def message_json(message)
          {
            id: message.id,
            room_id: message.room_id,
            creator_id: message.creator_id,
            creator_name: message.creator.name,
            body: message.plain_text_body,
            created_at: message.created_at.iso8601
          }
        end
    end
  end
end
