module Api
  module V1
    class RoomsController < BaseController
      before_action :set_room, only: %i[ show update destroy ]

      def index
        rooms = @current_api_user.rooms
        render json: rooms.map { |r| room_json(r) }
      end

      def show
        render json: room_json(@room)
      end

      def create
        if Current.account.settings.restrict_room_creation_to_administrators? && !@current_api_user.administrator?
          render json: { error: "Only administrators can create rooms" }, status: :forbidden
          return
        end

        room_type = params[:type].to_s.downcase
        case room_type
        when "closed"
          users = User.where(id: Array(params[:user_ids]))
          users = users.or(User.where(id: @current_api_user.id)) unless users.exists?(id: @current_api_user.id)
          room = Rooms::Closed.create_for(room_params, users: users)
        else
          room = Rooms::Open.create_for(room_params, users: @current_api_user)
        end

        render json: room_json(room), status: :created
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      def update
        unless @current_api_user.can_administer?(@room)
          render json: { error: "Not authorized" }, status: :forbidden
          return
        end

        @room.update!(room_params)

        if params[:user_ids].present? && @room.is_a?(Rooms::Closed)
          grantees = User.where(id: Array(params[:user_ids]))
          revokees = @room.users.where.not(id: Array(params[:user_ids]))
          @room.memberships.revise(granted: grantees, revoked: revokees)
        end

        render json: room_json(@room.reload)
      rescue ActiveRecord::RecordInvalid => e
        render json: { error: e.message }, status: :unprocessable_entity
      end

      def destroy
        unless @current_api_user.can_administer?(@room)
          render json: { error: "Not authorized" }, status: :forbidden
          return
        end

        @room.destroy
        head :no_content
      end

      # POST /api/v1/rooms/direct — find or create a DM room with another user.
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
        def set_room
          @room = @current_api_user.rooms.find(params[:id])
        rescue ActiveRecord::RecordNotFound
          render json: { error: "Room not found" }, status: :not_found
        end

        def room_params
          params.permit(:name)
        end

        def room_json(room)
          {
            id: room.id,
            name: room.name.presence || room.users.ordered.pluck(:name).join(", "),
            type: room.type,
            direct: room.direct?,
            member_ids: room.memberships.pluck(:user_id)
          }
        end
    end
  end
end
