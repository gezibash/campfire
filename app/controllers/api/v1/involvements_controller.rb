module Api
  module V1
    class InvolvementsController < BaseController
      def update
        membership = @current_api_user.memberships.find_by!(room_id: params[:id])
        membership.update!(involvement: params[:involvement])

        render json: {
          room_id: membership.room_id,
          involvement: membership.involvement
        }
      rescue ActiveRecord::RecordNotFound
        render json: { error: "Room not found or not a member" }, status: :not_found
      rescue ArgumentError => e
        render json: { error: e.message }, status: :unprocessable_entity
      end
    end
  end
end
