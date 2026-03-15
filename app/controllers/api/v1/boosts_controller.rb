module Api
  module V1
    class BoostsController < BaseController
      def create
        message = @current_api_user.reachable_messages.find(params[:message_id])

        boost = message.boosts.create!(content: params[:content])

        broadcast_boosts(message)

        render json: { id: boost.id, content: boost.content, booster: @current_api_user.name }, status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      def destroy
        boost = @current_api_user.boosts.find(params[:id])
        message = boost.message
        boost.destroy!

        broadcast_boosts(message)

        head :no_content
      rescue ActiveRecord::RecordNotFound
        render json: { error: "Boost not found" }, status: :not_found
      end

      private
        def broadcast_boosts(message)
          message.reload
          Turbo::StreamsChannel.broadcast_replace_to(
            [ message.room, :messages ],
            target: ActionView::RecordIdentifier.dom_id(message, :boosting),
            partial: "messages/boosts/boosts",
            locals: { message: message },
            attributes: { maintain_scroll: true }
          )
        end
    end
  end
end
