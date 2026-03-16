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
      mentionee_ids = event == :message_created ? mentionees.pluck(:id).to_set : Set.new

      base_payload = {
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
          direct: room.direct?
        }
      }

      # Broadcast to each room member's personal API stream with per-user mentioned flag.
      member_ids.each do |uid|
        payload = base_payload.deep_dup
        payload[:mentioned] = mentionee_ids.include?(uid)
        ActionCable.server.broadcast("api:user:#{uid}", payload)
      end
    end
end
