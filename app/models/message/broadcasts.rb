module Message::Broadcasts
  def broadcast_create
    broadcast_append_to room, :messages, target: [ room, :messages ]
    ActionCable.server.broadcast("unread_rooms", { roomId: room.id })
    broadcast_api_event(:message_created)
  end

  def broadcast_remove
    broadcast_remove_to room, :messages
    broadcast_api_event(:message_removed)
  end

  private
    def broadcast_api_event(event)
      member_ids = room.memberships.pluck(:user_id)

      payload = {
        event: event,
        message: {
          id: id,
          room_id: room_id,
          creator_id: creator_id,
          creator_name: creator.name,
          body: plain_text_body,
          created_at: created_at.iso8601
        },
        room: {
          id: room.id,
          name: room.name.presence || room.users.ordered.pluck(:name).join(", "),
          type: room.type,
          direct: room.direct?,
          member_ids: member_ids
        }
      }

      # Broadcast to each room member's personal API stream.
      member_ids.each do |uid|
        ActionCable.server.broadcast("api:user:#{uid}", payload)
      end

      # Broadcast to the global API stream so admin watchers see all messages.
      ActionCable.server.broadcast("api:global", payload)
    end
end
