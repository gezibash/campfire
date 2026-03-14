module Api
  module V1
    class RoomsController < BaseController
      def index
        rooms = @current_api_user.rooms
        render json: rooms.map { |r| room_json(r) }
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
