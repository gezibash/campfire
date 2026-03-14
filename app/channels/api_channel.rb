class ApiChannel < ApplicationCable::Channel
  def subscribed
    stream_from "api:user:#{current_user.id}"
    stream_from "api:global"
  end
end
