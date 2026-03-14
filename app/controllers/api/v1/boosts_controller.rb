module Api
  module V1
    class BoostsController < BaseController
      def create
        message = @current_api_user.reachable_messages.find(params[:message_id])

        Current.user = @current_api_user
        boost = message.boosts.create!(content: params[:content])

        boost.broadcast_append_to boost.message.room, :messages,
          target: "boosts_message_#{boost.message.client_message_id}",
          partial: "messages/boosts/boost",
          attributes: { maintain_scroll: true }

        render json: { id: boost.id, content: boost.content, booster: @current_api_user.name }, status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end
    end
  end
end
