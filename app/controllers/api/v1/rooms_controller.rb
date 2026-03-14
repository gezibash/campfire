module Api
  module V1
    class RoomsController < BaseController
      def index
        rooms = @current_api_user.rooms
        render json: rooms.map { |r| room_json(r) }
      end

      # POST /api/v1/rooms/direct — find or create a DM room with another user.
      # Params: user_id (the other user's ID)
      def direct
        other = User.active.find_by(id: params[:user_id])
        unless other
          render json: { error: "User not found" }, status: :not_found
          return
        end

        room = Rooms::Direct.find_or_create_for(User.where(id: [ @current_api_user.id, other.id ]))
        render json: room_json(room), status: :ok
      end

      private
        def room_json(room)
          {
            id: room.id,
            name: room.name.presence || room.users.ordered.pluck(:name).join(", "),
            type: room.type,
            direct: room.direct?
          }
        end
    end
  end
end
