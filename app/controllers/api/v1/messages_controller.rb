module Api
  module V1
    class MessagesController < BaseController
      include ActiveStorage::SetCurrent

      def index
        room = @current_api_user.rooms.find(params[:room_id])

        if params[:around].present?
          center = room.messages.with_creator.find(params[:around])
          half = params.fetch(:limit, 5).to_i.clamp(1, 50)
          before_msgs = room.messages.with_creator.ordered.where("id < ?", center.id).last(half)
          after_msgs  = room.messages.with_creator.ordered.where("id > ?", center.id).limit(half)
          messages = before_msgs + [center] + after_msgs
          render json: messages.map { |m| message_json(m) }
          return
        end

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

        message = room.messages.create!(body: params[:body])
        message.broadcast_create

        render json: message_json(message), status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      def destroy
        room = @current_api_user.rooms.find(params[:room_id])
        message = room.messages.find(params[:id])

        unless @current_api_user.can_administer?(message)
          render json: { error: "Not authorized" }, status: :forbidden
          return
        end

        message.destroy
        message.broadcast_remove
        head :no_content
      rescue ActiveRecord::RecordNotFound
        render json: { error: "Message not found" }, status: :not_found
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
