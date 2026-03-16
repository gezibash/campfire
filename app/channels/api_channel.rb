class ApiChannel < ApplicationCable::Channel
  def subscribed
    stream_from "api:user:#{current_user.id}"
  end
end
