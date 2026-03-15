module Api
  module V1
    class SearchesController < BaseController
      def index
        query = params[:q].to_s.gsub(/[^[:word:]]/, " ").strip

        if query.blank?
          render json: { error: "Query parameter 'q' is required" }, status: :unprocessable_entity
          return
        end

        limit = params.fetch(:limit, 100).to_i.clamp(1, 200)
        messages = @current_api_user.reachable_messages.search(query).last(limit)

        render json: messages.map { |m| search_result_json(m) }
      end

      private
        def search_result_json(message)
          {
            id: message.id,
            room_id: message.room_id,
            room_name: message.room.name.presence || message.room.users.ordered.pluck(:name).join(", "),
            creator_id: message.creator_id,
            creator_name: message.creator.name,
            body: message.plain_text_body,
            created_at: message.created_at.iso8601
          }
        end
    end
  end
end
